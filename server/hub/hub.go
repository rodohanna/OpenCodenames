package hub

import (
	"context"
	"errors"
	"log"

	"../db"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/websocket"
)

// IncomingMessage represents actions players send to the server.
type IncomingMessage struct {
	Action string
}

// Client represents a player or spectator
type Client struct {
	GameID        string
	PlayerID      string
	Hub           *Hub
	Conn          *websocket.Conn
	Incoming      chan *IncomingMessage
	Cancel        chan struct{}
	SpectatorOnly bool
	send          chan *db.Game
}

// NewClient creates a new client
func NewClient(gameID string, playerID string, hub *Hub, conn *websocket.Conn, spectator bool) *Client {
	return &Client{
		GameID:        gameID,
		PlayerID:      playerID,
		Hub:           hub,
		Conn:          conn,
		Incoming:      make(chan *IncomingMessage),
		Cancel:        make(chan struct{}),
		SpectatorOnly: spectator,
		send:          make(chan *db.Game),
	}
}

type spectatorGame struct {
	ID      string
	Status  string
	Players []string
	You     string
}

func mapGameToSpectatorGame(game *db.Game, playerID string) (*spectatorGame, error) {
	if _, ok := game.Players[playerID]; !ok {
		return nil, errors.New("Provided game and playerID do not match")
	}
	log.Println(game.Players)
	sg := &spectatorGame{
		ID:      game.ID,
		Status:  game.Status,
		Players: make([]string, 0, len(game.Players)),
		You:     game.Players[playerID]}
	for _, playerName := range game.Players {
		sg.Players = append(sg.Players, playerName)
	}
	return sg, nil
}

// Listen broadcasts game changes and handles client actions
func (c *Client) Listen() {
	for {
		select {
		case message := <-c.Incoming:
			if c.SpectatorOnly {
				log.Println("only a specator, limited abilities")
				continue
			}
			log.Println("recv", message)
		case game := <-c.send:
			log.Println("send", game)
			sg, err := mapGameToSpectatorGame(game, c.PlayerID)
			if err != nil {
				log.Println("Game broadcast error", err)
			}
			c.Conn.WriteJSON(map[string]*spectatorGame{"game": sg})
		case <-c.Cancel:
			log.Println("done")
			c.Hub.unregister <- c
			return
		}

	}
}

// Hub manages clients and connections by game
type Hub struct {
	clients         map[string]map[string]*Client // map of gameID to [map of PlayerID to Client]
	games           map[string]*db.Game           // map of gameID to Game
	watchers        map[string]*GameWatcher
	fireStoreClient *firestore.Client
	gameBroadcast   chan *db.Game
	Register        chan *Client
	unregister      chan *Client
}

// NewHub creates a new hub
func NewHub(client *firestore.Client) *Hub {
	return &Hub{
		clients:         map[string]map[string]*Client{},
		games:           map[string]*db.Game{},
		watchers:        map[string]*GameWatcher{},
		fireStoreClient: client,
		Register:        make(chan *Client),
		gameBroadcast:   make(chan *db.Game),
		unregister:      make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	defer func() {
		close(h.gameBroadcast)
	}()
	ctx := context.Background()
	for {
		select {
		// When a game changes, messages are pushed onto this channel to be broadcasted to
		// all participants
		case game := <-h.gameBroadcast:
			log.Println("broadcasting game change", h.games, h.clients)
			h.games[game.ID] = game
			for _, client := range h.clients[game.ID] {
				client.send <- game
			}
		// When a client wants to join a game they push themselves onto this channel
		case client := <-h.Register:
			log.Println("client registration", client)
			game, err := db.GetGame(ctx, h.fireStoreClient, client.GameID)
			if err != nil {
				log.Println("Could not find game", err)
				client.Conn.WriteJSON(map[string]string{"error": "could not find game"})
				// closing the connection will trigger the client cancellation process from SpectatorHandler & PlayerHandler
				client.Conn.Close()
				continue
			}
			if _, ok := game.Players[client.PlayerID]; !ok && !client.SpectatorOnly {
				log.Println("Player does not belong to game and is not spectator", err)
				client.Conn.WriteJSON(map[string]string{"error": "access denied"})
				// closing the connection will trigger the client cancellation process from SpectatorHandler & PlayerHandler
				client.Conn.Close()
				continue
			}
			if h.clients[game.ID] == nil {
				h.clients[game.ID] = make(map[string]*Client)
			}
			h.clients[game.ID][client.PlayerID] = client
			client.send <- game
			// If there is no game watcher set up we need to start one
			if _, ok := h.watchers[game.ID]; !ok {
				h.games[game.ID] = game
				h.watchers[game.ID] = &GameWatcher{game.ID, h.gameBroadcast, make(chan struct{})}
				go h.watchers[game.ID].watch(h.fireStoreClient)
			}
			log.Println("we registered")
		// When a client leaves a game or we decide to close the connection
		case client := <-h.unregister:
			log.Println("client unregistration", client)
			client.Conn.Close()
			close(client.send)
			close(client.Incoming)
			if _, ok := h.clients[client.GameID][client.PlayerID]; ok {
				log.Println("removing client from hub")
				delete(h.clients[client.GameID], client.PlayerID)
				if len(h.clients[client.GameID]) == 0 {
					log.Println("no more clients, stopping game watcher")
					close(h.watchers[client.GameID].cancel)
					delete(h.watchers, client.GameID)

				}
			}
		}
	}
}

// GameWatcher listens for changes on a game and sends them to the gameBroadcast channel
type GameWatcher struct {
	gameID        string
	gameBroadcast chan *db.Game
	cancel        chan struct{}
}

func (gw *GameWatcher) watch(client *firestore.Client) {
	log.Println("watching game", gw.gameID)
	ctx := context.Background()
	stop := db.ListenToGame(ctx, client, gw.gameID, func(game *db.Game) {
		gw.gameBroadcast <- game
		log.Println("game update", game)
	})
	for {
		select {
		case <-gw.cancel:
			stop()
			log.Println("cancelling watcher", gw.gameID)
			return
		}
	}
}

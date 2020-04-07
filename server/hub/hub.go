package hub

import (
	"context"
	"log"

	"../db"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/websocket"
)

// IncomingMessage todo
type IncomingMessage struct {
	action string
}

// Client todo
type Client struct {
	GameID   string
	PlayerID string
	Hub      *Hub
	Conn     *websocket.Conn
	Incoming chan *IncomingMessage
	Cancel   chan struct{}
	send     chan *db.Game
}

// NewClient todo
func NewClient(gameID string, playerID string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{GameID: gameID, PlayerID: playerID, Hub: hub, Conn: conn, Incoming: make(chan *IncomingMessage), send: make(chan *db.Game), Cancel: make(chan struct{})}
}

// Listen todo
func (c *Client) Listen() {
	for {
		select {
		case message := <-c.Incoming:
			log.Println("recv", message)
		case game := <-c.send:
			log.Println("send", game)
			c.Conn.WriteJSON(map[string]db.Game{"game": *game})
		case <-c.Cancel:
			log.Println("done")
			c.Hub.unregister <- c
			return
		}

	}
}

// Hub todo
type Hub struct {
	clients         map[string]map[string]*Client // map of gameID to [map of PlayerID to Client]
	games           map[string]*db.Game           // map of gameID to Game
	watchers        map[string]*GameWatcher
	fireStoreClient *firestore.Client
	gameBroadcast   chan *db.Game
	Register        chan *Client
	unregister      chan *Client
}

// NewHub todo
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

// Run todo
func (h *Hub) Run() {
	defer func() {
		close(h.gameBroadcast)
	}()
	ctx := context.Background()
	for {
		select {
		case game := <-h.gameBroadcast:
			log.Println("game broadcast", h.games, h.clients)
			if _, ok := h.games[game.ID]; ok {
				h.games[game.ID] = game
				for _, client := range h.clients[game.ID] {
					client.send <- game
				}
			}
		case client := <-h.Register:
			log.Println("new client", client)
			game, err := db.GetGame(ctx, h.fireStoreClient, client.GameID)
			if err != nil {
				log.Println("Could not find game", err)
				return
			}
			if _, ok := game.Players[client.PlayerID]; !ok {
				log.Println("Player does not belong to game", err)
				return
			}
			if h.clients[game.ID] == nil {
				h.clients[game.ID] = make(map[string]*Client)
			}
			h.clients[game.ID][client.PlayerID] = client
			client.send <- game
			if _, ok := h.watchers[game.ID]; !ok {
				h.games[game.ID] = game
				h.watchers[game.ID] = &GameWatcher{game.ID, h.gameBroadcast, make(chan struct{})}
				go h.watchers[game.ID].watch(h.fireStoreClient)
			}
			log.Println("we registered")
		case client := <-h.unregister:
			log.Println("old client", client)
			client.Conn.Close()
			close(client.send)
			close(client.Incoming)
			if _, ok := h.clients[client.GameID][client.PlayerID]; ok {
				log.Println("cleaning up")
				delete(h.clients[client.GameID], client.PlayerID)
				if len(h.clients[client.GameID]) == 0 {
					log.Println("should stop watcher...")
					close(h.watchers[client.GameID].cancel)
					delete(h.watchers, client.GameID)

				}
			}
		}
	}
}

// GameWatcher todo
type GameWatcher struct {
	gameID        string
	gameBroadcast chan *db.Game
	cancel        chan struct{}
}

func (gw *GameWatcher) watch(client *firestore.Client) {
	log.Println("watching a game", gw.gameID)
	ctx := context.Background()
	stop := db.ListenToGame(ctx, client, gw.gameID, func(game *db.Game) {
		gw.gameBroadcast <- game
		log.Println("game update", game)
	})
	for {
		select {
		case <-gw.cancel:
			stop()
			log.Println("done.....")
			return
		}
	}
}

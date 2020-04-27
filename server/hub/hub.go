package hub

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/RobertDHanna/OpenCodenames/db"
	g "github.com/RobertDHanna/OpenCodenames/game"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// IncomingMessage represents actions players send to the server.
type IncomingMessage struct {
	Action string
}

// Client represents a player or spectator
type Client struct {
	GameID        string
	PlayerID      string
	SessionID     string
	Hub           *Hub
	Conn          *websocket.Conn
	Cancel        chan struct{}
	SpectatorOnly bool
	send          chan *db.Game
}

// NewClient creates a new client
func NewClient(gameID string, playerID string, sessionID string, hub *Hub, conn *websocket.Conn, spectator bool) *Client {
	return &Client{
		GameID:        gameID,
		PlayerID:      playerID,
		SessionID:     sessionID,
		Hub:           hub,
		Conn:          conn,
		Cancel:        make(chan struct{}),
		SpectatorOnly: spectator,
		send:          make(chan *db.Game),
	}
}

func broadcastGame(c *Client, game *db.Game) error {
	c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	w, err := c.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	send := func(w io.WriteCloser, thing interface{}) error {
		j, err := json.Marshal(thing)
		if err != nil {
			return err
		}
		if _, err := w.Write(j); err != nil {
			return err
		}
		return nil
	}
	if c.SpectatorOnly {
		bg, err := g.MapGameToBaseGame(game)
		if err != nil {
			log.Println("MapGameToBaseGame error", err)
		}
		if err := send(w, g.PlayerGame{BaseGame: *bg}); err != nil {
			return err
		}
	} else {
		playerName := game.Players[c.PlayerID]
		if game.TeamRedSpy == playerName || game.TeamBlueSpy == playerName {
			sg, err := g.MapGameToSpyGame(game, c.PlayerID)
			if err != nil {
				log.Println("MapGameToSpyGame error", err)
			}
			if err := send(w, sg); err != nil {
				return err
			}
		} else {
			gg, err := g.MapGameToGuesserGame(game, c.PlayerID)
			if err != nil {
				log.Println("MapGameToGuesserGame error", err)
			}
			if err := send(w, gg); err != nil {
				return err
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	ctx := context.Background()
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var message IncomingMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			log.Println("Dropping connection, client encountered error", err)
			break
		}
		if c.SpectatorOnly {
			log.Println("Spectator attempted action:", message)
			continue
		}
		switch {
		case message.Action == "StartGame":
			log.Println("StartGame Handler")
			game, ok := c.Hub.games[c.GameID]
			if !ok {
				log.Println("Error: could not find client game")
				continue
			}
			g.HandleGameStart(ctx, c.Hub.fireStoreClient, game, c.PlayerID)
		case strings.Contains(message.Action, "Guess"):
			game := c.Hub.games[c.GameID]
			g.HandlePlayerGuess(ctx, c.Hub.fireStoreClient, message.Action, c.PlayerID, game)
			log.Println("Guess ended")
		case message.Action == "EndTurn":
			game := c.Hub.games[c.GameID]
			g.HandleEndTurn(ctx, c.Hub.fireStoreClient, game, c.PlayerID)
		case strings.Contains(message.Action, "UpdateTeam"):
			log.Println("UpdateTeam Handler")
			game := c.Hub.games[c.GameID]
			g.HandleUpdateTeams(ctx, c.Hub.fireStoreClient, game, message.Action, c.PlayerID)
		}
		log.Println("Received: ", message)
	}
}

// WritePump consumes messages off of the send channel and send them to the client
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case game, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := broadcastGame(c, game)
			if err != nil {
				log.Println("broadcaseGame err:", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub manages clients and connections by game
type Hub struct {
	clients         map[string]map[string]*Client // map of gameID to [map of PlayerID to Client]
	games           map[string]*db.Game           // map of gameID to Game
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
		fireStoreClient: client,
		Register:        make(chan *Client),
		gameBroadcast:   make(chan *db.Game),
		unregister:      make(chan *Client),
	}
}

func reapClient(client *Client, hub *Hub) {
	if _, ok := hub.clients[client.GameID][client.SessionID]; ok {
		log.Println("Removing client from hub")
		close(client.send)
		delete(hub.clients[client.GameID], client.SessionID)
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
			log.Println("Broadcasting game change", h.games, h.clients)
			h.games[game.ID] = game
			for _, client := range h.clients[game.ID] {
				select {
				case client.send <- game:
				default:
					log.Println("Closing client to not block")
					reapClient(client, h)
				}
			}
		// When a client wants to join a game they push themselves onto this channel
		case client := <-h.Register:
			log.Println("Client registration", client)
			game, err := db.GetGame(ctx, h.fireStoreClient, client.GameID)
			if err != nil {
				log.Println("Could not find game", err)
				client.Conn.WriteJSON(map[string]string{"error": "could not find game"})
				reapClient(client, h)
				continue
			}
			if _, ok := game.Players[client.PlayerID]; !ok && !client.SpectatorOnly {
				log.Println("Player does not belong to game and is not spectator", err)
				client.Conn.WriteJSON(map[string]string{"error": "access denied"})
				reapClient(client, h)
				continue
			}
			if h.clients[game.ID] == nil {
				h.clients[game.ID] = make(map[string]*Client)
			}
			h.clients[game.ID][client.SessionID] = client
			client.send <- game
			h.games[game.ID] = game
			log.Println("Finished client registration")
		// When a client leaves a game or we decide to close the connection
		case client := <-h.unregister:
			log.Println("Client unregistration", client)
			reapClient(client, h)
		}
	}
}

// ListenToGames listens for any changes on any games and broadcasts it via gameBroadcast
func (h *Hub) ListenToGames() {
	ctx := context.Background()
	iter := db.ListenToGames(ctx, h.fireStoreClient)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		log.Println("looking at a doc", doc)
		if err != nil {
			log.Println("err", err)
			continue
		}
		for _, change := range doc.Changes {
			switch change.Kind {
			case firestore.DocumentModified:
				var game db.Game
				if err := change.Doc.DataTo(&game); err != nil {
					log.Println("Doc to game err", err)
					continue
				}
				log.Println("broadcasting game change listener")
				h.gameBroadcast <- &game
			case firestore.DocumentRemoved:
				continue
			}
		}
	}
}

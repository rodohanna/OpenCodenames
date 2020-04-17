package hub

import (
	"context"
	"errors"
	"log"
	"math/rand"

	"../data"
	"../db"
	"../utils"
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
	ID              string
	Status          string
	Players         []string
	You             string
	YourTurn        bool
	TeamRed         []string
	TeamBlue        []string
	TeamRedSpy      string
	TeamBlueSpy     string
	TeamRedGuesser  string
	TeamBlueGuesser string
	WhoseTurn       string
	Cards           map[string]db.Card
}

func mapGameToSpectatorGame(game *db.Game, playerID string) (*spectatorGame, error) {
	if _, ok := game.Players[playerID]; !ok {
		return nil, errors.New("Provided game and playerID do not match")
	}
	_, belongsToTeamRed := game.TeamRed[playerID]
	_, belongsToTeamBlue := game.TeamBlue[playerID]
	if game.Status == "running" && !belongsToTeamRed && !belongsToTeamBlue {
		return nil, errors.New("PlayerID doesn't belong to game")
	}
	returnCards := map[string]db.Card{}
	for word, card := range game.Cards {
		returnCard := db.Card{BelongsTo: "", Guessed: card.Guessed, Index: card.Index}
		if card.Guessed {
			returnCard.BelongsTo = card.BelongsTo
			returnCard.Guessed = true
		}
		returnCards[word] = returnCard
	}
	sg := &spectatorGame{
		ID:              game.ID,
		Status:          game.Status,
		Players:         make([]string, 0, len(game.Players)),
		You:             game.Players[playerID],
		YourTurn:        false,
		TeamRed:         make([]string, 0, len(game.TeamRed)),
		TeamBlue:        make([]string, 0, len(game.TeamBlue)),
		TeamRedSpy:      game.TeamRedSpy,
		TeamBlueSpy:     game.TeamBlueSpy,
		Cards:           returnCards,
		TeamRedGuesser:  "",
		TeamBlueGuesser: "",
		WhoseTurn:       game.WhoseTurn}
	if _, ok := game.TeamRed[playerID]; ok && game.WhoseTurn == "red" {
		sg.YourTurn = true
	} else if _, ok := game.TeamBlue[playerID]; ok && game.WhoseTurn == "blue" {
		sg.YourTurn = true
	}
	for _, playerName := range game.Players {
		sg.Players = append(sg.Players, playerName)
	}
	for _, playerName := range game.TeamRed {
		sg.TeamRed = append(sg.TeamRed, playerName)
	}
	for _, playerName := range game.TeamBlue {
		sg.TeamBlue = append(sg.TeamBlue, playerName)
	}
	if game.Status == "running" {
		sg.TeamRedGuesser = game.TeamRedGuesser
		sg.TeamBlueGuesser = game.TeamBlueGuesser
	}
	return sg, nil
}

// Listen broadcasts game changes and handles client actions
func (c *Client) Listen() {
	ctx := context.Background()
	for {
		select {
		case message := <-c.Incoming:
			if c.SpectatorOnly {
				log.Println("only a specator, limited abilities")
				continue
			}
			switch message.Action {
			case "StartGame":
				log.Println("StartGame Handler")
				game, ok := c.Hub.games[c.GameID]
				if !ok {
					log.Println("Error: could not find client game")
					continue
				}
				if game.Status == "pending" && len(game.Players) >= 4 && game.CreatorID == c.PlayerID {
					log.Println("Starting Game", c.GameID)
					teamRed := map[string]string{}
					teamBlue := map[string]string{}
					teamRedIDs := make([]string, 0)
					teamBlueIDs := make([]string, 0)
					i := 0
					for playerID, playerName := range game.Players {
						if i%2 == 0 {
							teamRed[playerID] = playerName
							teamRedIDs = append(teamRedIDs, playerID)
						} else {
							teamBlue[playerID] = playerName
							teamBlueIDs = append(teamBlueIDs, playerID)
						}
						i++
					}
					teamRedSpyID := teamRedIDs[rand.Intn(len(teamRedIDs))]
					teamBlueSpyID := teamBlueIDs[rand.Intn(len(teamBlueIDs))]
					teamRedGuesserID := ""
					teamBlueGuesserID := ""

					for {
						teamRedGuesserID = teamRedIDs[rand.Intn(len(teamRedIDs))]
						if teamRedGuesserID == teamRedSpyID {
							continue
						}
						break
					}
					for {
						teamBlueGuesserID = teamBlueIDs[rand.Intn(len(teamBlueIDs))]
						if teamBlueGuesserID == teamBlueSpyID {
							continue
						}
						break
					}

					wordList := data.GetWordList()
					chosenWords := make([]string, 0, 25)
					for {
						randomWord := wordList[rand.Intn(len(wordList))]
						if _, contains := utils.Contains(chosenWords, randomWord); contains {
							continue
						}
						chosenWords = append(chosenWords, randomWord)
						if len(chosenWords) == 25 {
							break
						}
					}
					cards := map[string]db.Card{}
					i = 0
					for _, word := range chosenWords {
						cards[word] = db.Card{BelongsTo: "", Guessed: false, Index: i}
						i++
					}
					teamRedWords := make([]string, 0, 9)
					teamBlueWords := make([]string, 0, 8)
					// Select the bomb card
					blackWord := chosenWords[rand.Intn(len(chosenWords))]
					if card, ok := cards[blackWord]; ok {
						cards[blackWord] = db.Card{BelongsTo: "black", Guessed: false, Index: card.Index}
					}
					// Select red cards
					for j := 0; j < 8; j++ {
						randomWord := ""
						for {
							randomWord = chosenWords[rand.Intn(len(chosenWords))]
							if randomWord == blackWord {
								continue
							}
							if _, contains := utils.Contains(teamRedWords, randomWord); contains {
								continue
							}
							teamRedWords = append(teamRedWords, randomWord)
							break
						}
						if card, ok := cards[randomWord]; ok {
							cards[randomWord] = db.Card{BelongsTo: "red", Guessed: false, Index: card.Index}
						} else {
							log.Println("red not found", randomWord)
						}
					}
					// Select blue cards
					for j := 0; j < 9; j++ {
						randomWord := ""
						for {
							randomWord = chosenWords[rand.Intn(len(chosenWords))]
							if randomWord == blackWord {
								continue
							}
							if _, contains := utils.Contains(teamBlueWords, randomWord); contains {
								continue
							}
							if _, contains := utils.Contains(teamRedWords, randomWord); contains {
								continue
							}
							teamBlueWords = append(teamBlueWords, randomWord)
							break
						}
						if card, ok := cards[randomWord]; ok {
							cards[randomWord] = db.Card{BelongsTo: "blue", Guessed: false, Index: card.Index}
						} else {
							log.Println("blue not found", randomWord)
						}
					}

					db.UpdateGame(ctx, c.Hub.fireStoreClient, c.GameID, map[string]interface{}{
						"status":          "running",
						"teamRed":         teamRed,
						"teamBlue":        teamBlue,
						"teamRedSpy":      teamRed[teamRedSpyID],
						"teamBlueSpy":     teamBlue[teamBlueSpyID],
						"teamRedGuesser":  teamRed[teamRedGuesserID],
						"teamBlueGuesser": teamBlue[teamBlueGuesserID],
						"cards":           cards,
						"whoseTurn":       "blue",
					})
				}
			}
			log.Println("recv", message.Action)
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

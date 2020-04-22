package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"../db"
	h "../hub"
	"../utils"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/websocket"
)

// CreateGameHandler TODO: document
func CreateGameHandler(client *firestore.Client) utils.Handler {
	return utils.PostRequest(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		id, err := utils.MakeEasyID(4)
		if err != nil {
			log.Panic("Could not make an ID")
		}
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Panic("Could not parse URL")
		}
		playerID, playerIDErr := utils.GetQueryValue(&paramMap, "playerID")
		playerName, playerNameErr := utils.GetQueryValue(&paramMap, "playerName")
		playerMap := make(map[string]string)
		teamRed := make(map[string]string)
		teamBlue := make(map[string]string)
		creatorID := ""
		if len(playerID) > 0 && len(playerName) > 0 && playerIDErr == nil && playerNameErr == nil {
			playerMap[playerID] = playerName
			creatorID = playerID
		}
		game := db.Game{
			ID:                       id,
			Status:                   "pending",
			Players:                  playerMap,
			CreatorID:                creatorID,
			TeamRed:                  teamRed,
			TeamBlue:                 teamBlue,
			TeamRedSpy:               "",
			TeamBlueSpy:              "",
			TeamRedGuesser:           "",
			TeamBlueGuesser:          "",
			WhoseTurn:                "",
			Cards:                    make(map[string]db.Card),
			LastCardGuessed:          "",
			LastCardGuessedBy:        "",
			LastCardGuessedCorrectly: false,
		}
		err = db.CreateGame(ctx, client, &game)
		if err != nil {
			fmt.Fprintf(w, "failed to create game %s %s!", r.Method, id)
			return
		}
		fmt.Fprintf(w, `{"id":"%s"}`, id)
	})
}

// JoinGameHandler TODO: document
func JoinGameHandler(client *firestore.Client) utils.Handler {
	return utils.PostRequest(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Panic("Could not parse URL")
		}
		gameID, err := utils.GetQueryValue(&paramMap, "gameID")
		if err != nil {
			fmt.Fprintf(w, "Invalid gameID")
			return
		}
		playerName, err := utils.GetQueryValue(&paramMap, "playerName")
		if err != nil {
			fmt.Fprintf(w, "Invalid playerName")
			return
		}
		playerID, err := utils.GetQueryValue(&paramMap, "playerID")
		if err != nil {
			fmt.Fprintf(w, "Invalid playerID")
			return
		}

		err = db.AddPlayerToGame(ctx, client, gameID, playerID, playerName)
		if err != nil {
			log.Printf("Failed to add player %s to %s!", playerName, gameID)
			fmt.Fprintf(w, `{"error":"%s"}`, err)
			return
		}
		fmt.Fprintf(w, `{"success":true}`)
	})
}

// EchoHandler TODO: document
func EchoHandler() utils.Handler {
	return utils.WebSocketRequest(func(r *http.Request, c *websocket.Conn) {
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL")
			return
		}
		gameIDArray, gameIDExists := paramMap["gameID"]
		if !gameIDExists || len(gameIDArray) != 1 {
			c.WriteJSON(map[string]string{"error": "missing gameID field"})
			c.Close()
			return
		}
		playerIDArray, playerIDExists := paramMap["playerID"]
		if !playerIDExists || len(playerIDArray) != 1 {
			c.WriteJSON(map[string]string{"error": "missing playerID field"})
			c.Close()
			return
		}
		gameID := gameIDArray[0]
		playerID := playerIDArray[0]
		log.Printf("Success: gameID %s playerID %s", gameID, playerID)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})
}

// SpectatorHandler todo
func SpectatorHandler(client *firestore.Client, hub *h.Hub) utils.Handler {
	return utils.WebSocketRequest(func(r *http.Request, c *websocket.Conn) {
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL")
			return
		}
		gameID, err := utils.GetQueryValue(&paramMap, "gameID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing gameID field"})
			c.Close()
			return
		}
		id, err := utils.MakeEasyID(10)
		if err != nil {
			c.WriteJSON(map[string]string{"error": "could not generate temporary id"})
			c.Close()
			return
		}
		client := h.NewClient(gameID, id, hub, c, true)
		hub.Register <- client
		go func() {
			for {
				var incoming h.IncomingMessage
				err := c.ReadJSON(&incoming)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					log.Println("dropping connection, spectator encountered error", err)
					close(client.Cancel)
					return
				}
				// We drop anything the client sends us because they are only spectating
				log.Println("Dropped: ", incoming)
			}
		}()
		client.Listen()
	})
}

// PlayerHandler todo
func PlayerHandler(client *firestore.Client, hub *h.Hub) utils.Handler {
	return utils.WebSocketRequest(func(r *http.Request, c *websocket.Conn) {
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL")
			return
		}
		gameID, err := utils.GetQueryValue(&paramMap, "gameID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing gameID field"})
			c.Close()
			return
		}
		playerID, err := utils.GetQueryValue(&paramMap, "playerID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing playerID field"})
			c.Close()
			return
		}
		log.Printf("Success: gameID %s playerID %s", gameID, playerID)
		client := h.NewClient(gameID, playerID, hub, c, false)
		hub.Register <- client
		go func() {
			for {
				var incoming h.IncomingMessage
				err := c.ReadJSON(&incoming)
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					log.Println("dropping connection, client encountered error", err)
					close(client.Cancel)
					return
				}
				client.Incoming <- &incoming
				log.Println("Received: ", incoming)
			}
		}()
		client.Listen()
	})
}

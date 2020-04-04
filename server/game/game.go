package game

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"../db"
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
		game := db.Game{ID: id, Status: "pending", Players: make([]string, 0)}
		err = db.CreateGame(ctx, client, &game)
		if err != nil {
			fmt.Fprintf(w, "failed to create game %s %s!", r.Method, id)
			return
		}
		fmt.Fprintf(w, "successfully created game %s %s!", r.Method, id)
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
		gameID, gameIDExists := paramMap["gameID"]
		if !gameIDExists || len(gameID) != 1 {
			fmt.Fprintf(w, "Invalid gameID")
			return
		}
		playerName, playerNameExists := paramMap["playerName"]
		if !playerNameExists || len(playerName) != 1 {
			fmt.Fprintf(w, "Invalid playerName")
			return
		}
		err = db.AddPlayerToGame(ctx, client, gameID[0], playerName[0])
		if err != nil {
			fmt.Fprintf(w, "Failed to add player %s to %s!", playerName[0], gameID[0])
			return
		}
		fmt.Fprintf(w, "Successfully added player \"%s\" to %s!", playerName[0], gameID[0])
	})
}

// EchoHandler TODO: document
func EchoHandler() utils.Handler {
	return utils.WebSocketRequest(func(c *websocket.Conn) {
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

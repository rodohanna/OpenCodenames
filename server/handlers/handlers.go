package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"cloud.google.com/go/firestore"
	"github.com/RobertDHanna/OpenCodenames/data"
	"github.com/RobertDHanna/OpenCodenames/db"
	h "github.com/RobertDHanna/OpenCodenames/hub"
	"github.com/RobertDHanna/OpenCodenames/recaptcha"
	"github.com/RobertDHanna/OpenCodenames/utils"
	"github.com/gorilla/websocket"
)

// CreateGameHandler TODO: document
func CreateGameHandler(client *firestore.Client) utils.Handler {
	return utils.PostRequest(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		id, err := utils.MakeEasyID(5)
		if err != nil {
			log.Println("Could not make an ID", err)
			return
		}
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL", err)
			return
		}
		playerName, playerNameErr := utils.GetQueryValue(&paramMap, "playerName")
		recaptchaResponse, recaptchaErr := utils.GetQueryValue(&paramMap, "recaptcha")
		if recaptchaResponse == "" || recaptchaErr != nil {
			log.Println("A ReCAPTCHA token is required")
			return
		}
		recaptcha.Init(data.GetReCAPTCHAKey())
		response, err := recaptcha.Check(utils.GetIP(r), recaptchaResponse)
		log.Println("ReCAPTCHA response: ", response)
		if response.Score < 0.1 || err != nil {
			log.Println("ReCAPTCHA request failed", err)
			return
		}
		playerMap := make(map[string]string)
		teamRed := make(map[string]string)
		teamBlue := make(map[string]string)
		creatorID := ""
		teamBlueSpy := ""
		playerID, err := utils.MakeEasyID(15)
		if err != nil {
			log.Println("Failure creating playerID", err)
		}
		if len(playerName) > 0 && playerNameErr == nil {
			playerMap[playerID] = playerName
			teamBlue[playerID] = playerName
			creatorID = playerID
			teamBlueSpy = playerName
		}
		game := db.Game{
			ID:                       id,
			Status:                   "pending",
			Players:                  playerMap,
			CreatorID:                creatorID,
			TeamRed:                  teamRed,
			TeamBlue:                 teamBlue,
			TeamRedSpy:               "",
			TeamBlueSpy:              teamBlueSpy,
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
		fmt.Fprintf(w, `{"id":"%s","playerID":"%s"}`, id, playerID)
	})
}

// JoinGameHandler Handles adding a player to game
func JoinGameHandler(client *firestore.Client) utils.Handler {
	return utils.PostRequest(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL", err)
			return
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
		playerID, err := utils.MakeEasyID(15)
		if err != nil {
			log.Println("Failure creating playerID", err)
		}
		err = db.AddPlayerToGame(ctx, client, gameID, playerID, playerName)
		if err != nil {
			if err.Error() == "playerAlreadyAdded" {
				fmt.Fprintf(w, `{"success":true,"playerID":"%s"}`, playerID)
				return
			}
			log.Printf("Failed to add player %s to %s!", playerName, gameID)
			fmt.Fprintf(w, `{"error":"%s"}`, err)
			return
		}
		fmt.Fprintf(w, `{"success":true,"playerID":"%s"}`, playerID)
	})
}

// SpectatorHandler subscribes a "player" to a game without them having to be a player.
func SpectatorHandler(client *firestore.Client, hub *h.Hub) utils.Handler {
	return utils.WebSocketRequest(func(r *http.Request, c *websocket.Conn) {
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("SpectatorHandler: Could not parse URL", err)
			return
		}
		gameID, err := utils.GetQueryValue(&paramMap, "gameID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing gameID field"})
			c.Close()
			return
		}
		sessionID, err := utils.GetQueryValue(&paramMap, "sessionID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing sessionID field"})
			c.Close()
			return
		}
		id, err := utils.MakeEasyID(15)
		if err != nil {
			c.WriteJSON(map[string]string{"error": "could not generate temporary id"})
			c.Close()
			return
		}
		client := h.NewClient(gameID, id, sessionID, hub, c, true)
		hub.Register <- client
		go client.ReadPump()
		go client.WritePump()
	})
}

// PlayerHandler subscribes a player to a game.
func PlayerHandler(client *firestore.Client, hub *h.Hub) utils.Handler {
	return utils.WebSocketRequest(func(r *http.Request, c *websocket.Conn) {
		paramMap, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Println("Could not parse URL", err)
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
		sessionID, err := utils.GetQueryValue(&paramMap, "sessionID")
		if err != nil {
			c.WriteJSON(map[string]string{"error": "missing sessionID field"})
			c.Close()
			return
		}
		log.Printf("Success: gameID %s playerID %s sessionID %s", gameID, playerID, sessionID)
		client := h.NewClient(gameID, playerID, sessionID, hub, c, false)
		hub.Register <- client
		go client.ReadPump()
		go client.WritePump()
	})
}

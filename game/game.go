package game

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"../crud"
	"../utils"
	"cloud.google.com/go/firestore"
)

// CreateGame TODO: document
func CreateGame(client *firestore.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		id, err := utils.MakeEasyID(4)
		if err != nil {
			log.Panic("Could not make an ID")
		}
		game := crud.Game{ID: id, Status: "pending", Players: make([]string, 0)}
		err = crud.CreateGame(ctx, client, &game)
		if err != nil {
			fmt.Fprintf(w, "failed to create game %s %s!", r.Method, id)
			return
		}
		fmt.Fprintf(w, "successfully created game %s %s!", r.Method, id)
	}
}

// JoinGame TODO: document
func JoinGame(client *firestore.Client) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := utils.MakeEasyID(4)
		if err != nil {
			log.Panic("Could not make an ID")
		}
		fmt.Fprintf(w, "/game/join %s %s!", r.Method, id)
	}
}

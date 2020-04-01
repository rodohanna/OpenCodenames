package game

import (
	"fmt"
	"log"
	"net/http"

	"../utils"
)

// CreateGame TODO: document
func CreateGame(w http.ResponseWriter, r *http.Request) {
	id, err := utils.MakeEasyID(4)
	if err != nil {
		log.Panic("Could not make an ID")
	}
	fmt.Fprintf(w, "/game/create %s %s!", r.Method, id)
}

// JoinGame TODO: document
func JoinGame(w http.ResponseWriter, r *http.Request) {
	id, err := utils.MakeEasyID(4)
	if err != nil {
		log.Panic("Could not make an ID")
	}
	fmt.Fprintf(w, "/game/join %s %s!", r.Method, id)
}

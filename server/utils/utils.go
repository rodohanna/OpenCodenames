package utils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Handler yep
type Handler func(w http.ResponseWriter, r *http.Request)

// WSHandler yep
type WSHandler func(r *http.Request, c *websocket.Conn)

// MakeEasyID creates an ID composed of random alphabetic characters
func MakeEasyID(length int) (string, error) {
	if length < 1 {
		return "", errors.New("cannot pass in length < 1")
	}
	var id strings.Builder
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(alphabet))
		character := string(alphabet[randomIndex])
		id.WriteString(character)
	}
	return id.String(), nil
}

// PostRequest todo
func PostRequest(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprintf(w, "Invalid HTTP method")
			return
		}
		handler(w, r)
	}
}

// GetRequest todo
func GetRequest(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			fmt.Fprintf(w, "Invalid HTTP method")
			return
		}
		handler(w, r)
	}
}

// WebSocketRequest TODO: document
func WebSocketRequest(handle WSHandler) Handler {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: actually check the origin.
		return true
	}
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("WS request")
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		handle(r, c)
	}
}

// Contains asdf
func Contains(slice []string, target string) (int, bool) {
	for index, str := range slice {
		if str == target {
			return index, true
		}
	}
	return -1, false
}

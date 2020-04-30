package utils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Handler type for HTTP handlers
type Handler func(w http.ResponseWriter, r *http.Request)

// WSHandler type for WebSocket handlers
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

// PostRequest wraps a POST request handler
func PostRequest(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprintf(w, "Invalid HTTP method")
			return
		}
		handler(w, r)
	}
}

// GetRequest wraps a GET request handler
func GetRequest(handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			fmt.Fprintf(w, "Invalid HTTP method")
			return
		}
		handler(w, r)
	}
}

// WebSocketRequest wraps a WebSocket request handler
func WebSocketRequest(handle WSHandler) Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: actually check the origin.
		return true
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("WebSocket upgrade error:", err)
			return
		}
		handle(r, c)
	}
}

// Contains takes in a slice of strings and looks for a target
func Contains(slice []string, target string) (int, bool) {
	for index, str := range slice {
		if str == target {
			return index, true
		}
	}
	return -1, false
}

// GetQueryValue Tries to find the value for a given param
func GetQueryValue(params *url.Values, paramName string) (string, error) {
	paramValueArray, paramValueExists := (*params)[paramName]
	if !paramValueExists || len(paramValueArray) != 1 {
		return "", errors.New("Could not find param value")
	}
	return paramValueArray[0], nil
}

// GetIP implementation borrowed from https://golangcode.com/get-the-request-ip-addr/
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

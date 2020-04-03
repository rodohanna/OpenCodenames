package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Handler yep
type Handler func(w http.ResponseWriter, r *http.Request)

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

// Contains asdf
func Contains(slice []string, target string) (int, bool) {
	for index, str := range slice {
		if str == target {
			return index, true
		}
	}
	return -1, false
}

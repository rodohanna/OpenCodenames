package utils

import (
	"errors"
	"math/rand"
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

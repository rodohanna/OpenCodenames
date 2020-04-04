package db

import (
	"context"
	"errors"
	"log"

	"../utils"
	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Game represents a game.
type Game struct {
	ID      string   `firestore:"id"`
	Status  string   `firestore:"status"`
	Players []string `firestore:"players"`
}

// CreateGame Creates a game or returns an error if one already exists
func CreateGame(ctx context.Context, client *firestore.Client, game *Game) error {
	ref := client.Collection("games").Doc(game.ID)
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		_, err := tx.Get(ref)
		if err != nil && status.Code(err) != codes.NotFound {
			return err
		}
		return tx.Set(ref, game)
	})
	if err != nil {
		log.Printf("CreateGame: An error has occurred: %s", err)
	}
	return err
}

// AddPlayerToGame TODO: document (UNTESTED)
func AddPlayerToGame(ctx context.Context, client *firestore.Client, gameID string, playerName string) error {
	ref := client.Collection("games").Doc(gameID)
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		var game Game
		if err := doc.DataTo(&game); err != nil {
			return err
		}
		_, playerNameAlreadyAdded := utils.Contains(game.Players, playerName)
		if playerNameAlreadyAdded {
			return errors.New("playerName already added")
		}
		allPlayers := append(game.Players, playerName)
		return tx.Set(ref, map[string]interface{}{
			"players": allPlayers,
		}, firestore.MergeAll)
	})
	if err != nil {
		log.Printf("JoinGame: An error has occurred: %s", err)
	}
	return err
}

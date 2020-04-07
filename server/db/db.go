package db

import (
	"context"
	"errors"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Game represents a game.
type Game struct {
	ID      string            `firestore:"id"`
	Status  string            `firestore:"status"`
	Players map[string]string `firestore:"players"`
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

// AddPlayerToGame TODO: document
func AddPlayerToGame(ctx context.Context, client *firestore.Client, gameID string, playerID string, playerName string) error {
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
		_, playerFound := game.Players[playerID]
		if playerFound {
			return errors.New("playerID already added")
		}
		game.Players[playerID] = playerName
		return tx.Set(ref, map[string]interface{}{
			"players": game.Players,
		}, firestore.MergeAll)
	})
	if err != nil {
		log.Printf("JoinGame: An error has occurred: %s", err)
	}
	return err
}

// GetGame todo
func GetGame(ctx context.Context, client *firestore.Client, gameID string) (*Game, error) {
	doc, err := client.Collection("games").Doc(gameID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var game Game
	if err := doc.DataTo(&game); err != nil {
		return nil, err
	}
	return &game, nil
}

// ListenToGame TODO: document
func ListenToGame(ctx context.Context, client *firestore.Client, gameID string, handler func(game *Game)) func() {
	iter := client.Collection("games").Query.Where("id", "==", gameID).Snapshots(ctx)
	go func() {
		for {
			doc, err := iter.Next()
			log.Println("looking at a doc", doc)
			if err != nil {
				log.Println("err", err)
				return
			}
			for _, change := range doc.Changes {
				switch change.Kind {
				case firestore.DocumentModified:
					var game Game
					if err := change.Doc.DataTo(&game); err != nil {
						return
					}
					handler(&game)
				case firestore.DocumentRemoved:
					return
				}
			}
		}
	}()
	return func() { iter.Stop() }
}

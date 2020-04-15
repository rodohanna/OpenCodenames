package db

import (
	"context"
	"errors"
	"log"

	"../config"
	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Card represents metadata about a word on the board.
type Card struct {
	Index     int    `firestore:"index"`
	BelongsTo string `firestore:"belongsTo"`
	Guessed   bool   `firestore:"guessed"`
}

// Game represents a game.
type Game struct {
	ID                   string            `firestore:"id"`
	Status               string            `firestore:"status"`
	Players              map[string]string `firestore:"players"`
	CreatorID            string            `firestore:"creatorID"`
	TeamRed              map[string]string `firestore:"teamRed"`
	TeamBlue             map[string]string `firestore:"teamBlue"`
	TeamRedSpy           string            `firestore:"teamRedSpy"`
	TeamBlueSpy          string            `firestore:"teamBlueSpy"`
	TeamRedGuesserIndex  int               `firestore:"teamRedGuesserIndex"`
	TeamBlueGuesserIndex int               `firestore:"teamBlueGuesserIndex"`
	WhoseTurn            string            `firestore:"whoseTurn"`
	Cards                map[string]Card   `firestore:"cards"`
}

// UpdateGame Updates a game
func UpdateGame(ctx context.Context, client *firestore.Client, gameID string, fieldsToUpdate map[string]interface{}) error {
	ref := client.Collection("games").Doc(gameID)
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Set(ref, fieldsToUpdate, firestore.MergeAll)
	})
	if err != nil {
		log.Printf("UpdateGame: An error has occurred: %s", err)
	}
	return err
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
		if _, playerFound := game.Players[playerID]; playerFound {
			return errors.New("playerAlreadyAdded")
		}
		for _, otherPlayerName := range game.Players {
			if playerName == otherPlayerName {
				return errors.New("nameAlreadyTaken")
			}
		}
		if len(game.Players) >= config.PlayerLimit() {
			return errors.New("gameIsFull")
		}
		if game.Status != "pending" {
			return errors.New("gameAlreadyStarted")
		}
		fieldsToUpdate := map[string]interface{}{}
		if len(game.Players) == 0 {
			fieldsToUpdate["creatorID"] = playerID
		}
		game.Players[playerID] = playerName
		fieldsToUpdate["players"] = game.Players
		return tx.Set(ref, fieldsToUpdate, firestore.MergeAll)
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

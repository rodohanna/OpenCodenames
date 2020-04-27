package db

import (
	"context"
	"errors"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/RobertDHanna/OpenCodenames/config"
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
	ID                       string            `firestore:"id"`
	Status                   string            `firestore:"status"`
	Players                  map[string]string `firestore:"players"`
	CreatorID                string            `firestore:"creatorID"`
	TeamRed                  map[string]string `firestore:"teamRed"`
	TeamBlue                 map[string]string `firestore:"teamBlue"`
	TeamRedSpy               string            `firestore:"teamRedSpy"`
	TeamBlueSpy              string            `firestore:"teamBlueSpy"`
	TeamRedGuesser           string            `firestore:"teamRedGuesser"`
	TeamBlueGuesser          string            `firestore:"teamBlueGuesser"`
	WhoseTurn                string            `firestore:"whoseTurn"`
	Cards                    map[string]Card   `firestore:"cards"`
	LastCardGuessed          string            `firestore:"lastCardGuessed"`
	LastCardGuessedBy        string            `firestore:"lastCardGuessedBy"`
	LastCardGuessedCorrectly bool              `firestore:"lastCardGuessedCorrectly"`
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

// UpdateGameFirestoreUpdate takes in a firestore Update array and processes it.
func UpdateGameFirestoreUpdate(ctx context.Context, client *firestore.Client, gameID string, fieldsToUpdate []firestore.Update) error {
	ref := client.Collection("games").Doc(gameID)
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, fieldsToUpdate)
	})
	if err != nil {
		log.Printf("UpdateGameFirestoreUpdate: An error has occurred: %s", err)
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
			if game.Status == "pending" {
				// Overwrite player name
				game.Players[playerID] = playerName
				return tx.Set(ref, map[string]interface{}{
					"players": game.Players,
				}, firestore.MergeAll)
			}
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
			fieldsToUpdate["teamBlueSpy"] = playerName
			game.TeamBlue[playerID] = playerName
			fieldsToUpdate["teamBlue"] = game.TeamBlue
		} else {
			// Try to put player on a team and in a role...
			if game.TeamBlueSpy == "" {
				fieldsToUpdate["teamBlueSpy"] = playerName
				game.TeamBlue[playerID] = playerName
				fieldsToUpdate["teamBlue"] = game.TeamBlue
			} else if game.TeamBlueGuesser == "" {
				fieldsToUpdate["teamBlueGuesser"] = playerName
				game.TeamBlue[playerID] = playerName
				fieldsToUpdate["teamBlue"] = game.TeamBlue
			} else if game.TeamRedSpy == "" {
				fieldsToUpdate["teamRedSpy"] = playerName
				game.TeamRed[playerID] = playerName
				fieldsToUpdate["teamRed"] = game.TeamRed
			} else if game.TeamRedGuesser == "" {
				fieldsToUpdate["teamRedGuesser"] = playerName
				game.TeamRed[playerID] = playerName
				fieldsToUpdate["teamRed"] = game.TeamRed
			} else if len(game.TeamBlue) < len(game.TeamRed) {
				game.TeamBlue[playerID] = playerName
				fieldsToUpdate["teamBlue"] = game.TeamBlue
			} else {
				game.TeamRed[playerID] = playerName
				fieldsToUpdate["teamRed"] = game.TeamRed
			}
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

// ListenToGames returns an iterator that returns games that have been updated.
func ListenToGames(ctx context.Context, client *firestore.Client) *firestore.QuerySnapshotIterator {
	return client.Collection("games").Snapshots(ctx)
}

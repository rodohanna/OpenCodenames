package db

import (
	"context"
	"errors"
	"log"
	"time"

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

// Game represents a codenames game.
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
	UpdatedAt                int64             `firestore:"updatedAt"`
	TimesPlayed              int64             `firestore:"timesPlayed"`
}

// UpdateGame updates a game using a caller-provided mapOfUpdates.
func UpdateGame(ctx context.Context, client *firestore.Client, gameID string, mapOfUpdates map[string]interface{}) error {
	ref := client.Collection("games").Doc(gameID)
	fieldsToUpdate := []firestore.Update{}
	now := time.Now()
	mapOfUpdates["updatedAt"] = now.Unix()
	for key, value := range mapOfUpdates {
		fieldsToUpdate = append(fieldsToUpdate, firestore.Update{Path: key, Value: value})
	}
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, fieldsToUpdate)
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
		doc, err := tx.Get(ref)
		if err != nil && status.Code(err) != codes.NotFound {
			return err
		}
		if doc != nil && doc.Exists() {
			return errors.New("GameAlreadyExists")
		}
		now := time.Now()
		game.UpdatedAt = now.Unix()
		return tx.Set(ref, game)
	})
	if err != nil {
		log.Printf("CreateGame: An error has occurred: %s", err)
	}
	return err
}

// AddPlayerToGame Adds a player to a game if it still pending. It also attempts to set a role for the given player.
func AddPlayerToGame(ctx context.Context, client *firestore.Client, gameID string, playerID string, playerName string) error {
	ref := client.Collection("games").Doc(gameID)
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil && status.Code(err) != codes.NotFound {
			return err
		}
		if doc != nil && !doc.Exists() {
			return errors.New("GameDoesntExist")
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
			return errors.New("PlayerAlreadyAdded")
		}
		for _, otherPlayerName := range game.Players {
			if playerName == otherPlayerName {
				return errors.New("NameAlreadyTaken")
			}
		}
		if len(game.Players) >= config.PlayerLimit() {
			return errors.New("GameIsFull")
		}
		if game.Status != "pending" {
			return errors.New("GameAlreadyStarted")
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
		now := time.Now()
		fieldsToUpdate["updatedAt"] = now.Unix()
		return tx.Set(ref, fieldsToUpdate, firestore.MergeAll)
	})
	if err != nil {
		log.Printf("JoinGame: An error has occurred: %s", err)
	}
	return err
}

// GetGame Returns a Game struct.
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

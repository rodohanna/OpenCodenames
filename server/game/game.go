package game

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"strings"

	"../config"
	"../data"
	"../db"
	"../utils"
	"cloud.google.com/go/firestore"
)

// BaseGame collection of fields that every participant needs
type BaseGame struct {
	ID                       string
	Status                   string
	Players                  []string
	TeamRed                  []string
	TeamBlue                 []string
	TeamRedSpy               string
	TeamBlueSpy              string
	TeamRedGuesser           string
	TeamBlueGuesser          string
	WhoseTurn                string
	Cards                    map[string]db.Card
	LastCardGuessed          string
	LastCardGuessedBy        string
	LastCardGuessedCorrectly bool
}

// PlayerGame collection of fields that only players (not spectators) need
type PlayerGame struct {
	You          string
	YouOwnGame   bool
	YourTurn     bool
	GameCanStart bool
	BaseGame     BaseGame
}

func playerCanGuess(game *db.Game, playerID string) bool {
	if game == nil {
		return false
	}
	redPlayerName, playerOnTeamRed := game.TeamRed[playerID]
	bluePlayerName, playerOnTeamBlue := game.TeamBlue[playerID]
	return (playerOnTeamRed && game.WhoseTurn == "red" && game.TeamRedGuesser == redPlayerName) ||
		(playerOnTeamBlue && game.WhoseTurn == "blue" && game.TeamBlueGuesser == bluePlayerName)
}

func playerGuessedCardCorrectly(game *db.Game, card *db.Card, playerID string) bool {
	if game == nil || card == nil {
		return false
	}
	_, playerOnTeamRed := game.TeamRed[playerID]
	_, playerOnTeamBlue := game.TeamBlue[playerID]
	return (card.BelongsTo == "red" && playerOnTeamRed) || (card.BelongsTo == "blue" && playerOnTeamBlue)
}

func playerCanEndTurn(game *db.Game, playerID string) bool {
	if game == nil {
		return false
	}
	playerNameRed, playerOnTeamRed := game.TeamRed[playerID]
	playerNameBlue, playerOnTeamBlue := game.TeamBlue[playerID]
	return (playerOnTeamRed && game.TeamRedGuesser == playerNameRed && game.WhoseTurn == "red") ||
		(playerOnTeamBlue && game.TeamBlueGuesser == playerNameBlue && game.WhoseTurn == "blue")
}

// MapGameToBaseGame takes a db game and maps it to a BaseGame
func MapGameToBaseGame(game *db.Game) (*BaseGame, error) {
	if game == nil {
		return nil, errors.New("Received a nil game")
	}
	returnCards := map[string]db.Card{}
	for word, card := range game.Cards {
		returnCard := db.Card{BelongsTo: "", Guessed: card.Guessed, Index: card.Index}
		if card.Guessed {
			returnCard.BelongsTo = card.BelongsTo
			returnCard.Guessed = true
		}
		returnCards[word] = returnCard
	}
	baseGame := &BaseGame{
		ID:                       game.ID,
		Status:                   game.Status,
		Players:                  make([]string, 0, len(game.Players)),
		TeamRed:                  make([]string, 0, len(game.TeamRed)),
		TeamBlue:                 make([]string, 0, len(game.TeamBlue)),
		TeamRedSpy:               game.TeamRedSpy,
		TeamBlueSpy:              game.TeamBlueSpy,
		Cards:                    returnCards,
		TeamRedGuesser:           "",
		TeamBlueGuesser:          "",
		WhoseTurn:                game.WhoseTurn,
		LastCardGuessed:          game.LastCardGuessed,
		LastCardGuessedBy:        game.LastCardGuessedBy,
		LastCardGuessedCorrectly: game.LastCardGuessedCorrectly,
	}
	for _, playerName := range game.Players {
		baseGame.Players = append(baseGame.Players, playerName)
	}
	for _, playerName := range game.TeamRed {
		baseGame.TeamRed = append(baseGame.TeamRed, playerName)
	}
	for _, playerName := range game.TeamBlue {
		baseGame.TeamBlue = append(baseGame.TeamBlue, playerName)
	}
	if game.Status == "running" || game.Status == "redwon" || game.Status == "bluewon" {
		baseGame.TeamRedGuesser = game.TeamRedGuesser
		baseGame.TeamBlueGuesser = game.TeamBlueGuesser
	}
	return baseGame, nil
}

// MapGameToGuesserGame takes a db game and maps it to a PlayerGame w/ card.BelongsTo stripped out
func MapGameToGuesserGame(game *db.Game, playerID string) (*PlayerGame, error) {
	baseGame, err := MapGameToBaseGame(game)
	if err != nil {
		log.Println("MapGameToGuesserGame:MapGameToBaseGame failed", err)
	}
	guesserGame := &PlayerGame{
		You:          game.Players[playerID],
		YouOwnGame:   game.CreatorID == playerID,
		YourTurn:     false,
		GameCanStart: len(game.Players) >= 4 && len(game.Players) <= config.PlayerLimit(),
		BaseGame:     *baseGame,
	}
	if _, ok := game.TeamRed[playerID]; ok && game.WhoseTurn == "red" {
		guesserGame.YourTurn = true
	} else if _, ok := game.TeamBlue[playerID]; ok && game.WhoseTurn == "blue" {
		guesserGame.YourTurn = true
	}
	return guesserGame, nil
}

// MapGameToSpyGame takes a db game and maps it to a PlayerGame w/ card.BelongsTo still in
func MapGameToSpyGame(game *db.Game, playerID string) (*PlayerGame, error) {
	baseGame, err := MapGameToBaseGame(game)
	if err != nil {
		log.Println("MapGameToSpyGame:MapGameToBaseGame failed", err)
	}
	// Add BelongsTo back in
	for word, card := range game.Cards {
		baseGame.Cards[word] = card
	}
	spyGame := &PlayerGame{
		You:          game.Players[playerID],
		YouOwnGame:   game.CreatorID == playerID,
		YourTurn:     false,
		GameCanStart: len(game.Players) >= 4 && len(game.Players) <= config.PlayerLimit(),
		BaseGame:     *baseGame,
	}
	if _, ok := game.TeamRed[playerID]; ok && game.WhoseTurn == "red" {
		spyGame.YourTurn = true
	} else if _, ok := game.TeamBlue[playerID]; ok && game.WhoseTurn == "blue" {
		spyGame.YourTurn = true
	}
	return spyGame, nil
}

// HandleGameStart takes in a game and puts it into a "running" state
func HandleGameStart(ctx context.Context, client *firestore.Client, game *db.Game, playerID string) {
	if game.Status == "pending" && len(game.Players) >= 4 && game.CreatorID == playerID {
		log.Println("Starting Game", game.ID)
		teamRed := map[string]string{}
		teamBlue := map[string]string{}
		teamRedIDs := make([]string, 0)
		teamBlueIDs := make([]string, 0)
		i := 0
		for playerID, playerName := range game.Players {
			if i%2 == 0 {
				teamRed[playerID] = playerName
				teamRedIDs = append(teamRedIDs, playerID)
			} else {
				teamBlue[playerID] = playerName
				teamBlueIDs = append(teamBlueIDs, playerID)
			}
			i++
		}
		teamRedSpyID := teamRedIDs[rand.Intn(len(teamRedIDs))]
		teamBlueSpyID := teamBlueIDs[rand.Intn(len(teamBlueIDs))]
		teamRedGuesserID := ""
		teamBlueGuesserID := ""
		for {
			teamRedGuesserID = teamRedIDs[rand.Intn(len(teamRedIDs))]
			if teamRedGuesserID == teamRedSpyID {
				continue
			}
			break
		}
		for {
			teamBlueGuesserID = teamBlueIDs[rand.Intn(len(teamBlueIDs))]
			if teamBlueGuesserID == teamBlueSpyID {
				continue
			}
			break
		}

		wordList := data.GetWordList()
		chosenWords := make([]string, 0, 25)
		for {
			randomWord := wordList[rand.Intn(len(wordList))]
			if _, contains := utils.Contains(chosenWords, randomWord); contains {
				continue
			}
			chosenWords = append(chosenWords, randomWord)
			if len(chosenWords) == 25 {
				break
			}
		}
		cards := map[string]db.Card{}
		i = 0
		for _, word := range chosenWords {
			cards[word] = db.Card{BelongsTo: "", Guessed: false, Index: i}
			i++
		}
		teamRedWords := make([]string, 0, 9)
		teamBlueWords := make([]string, 0, 8)
		// Select the bomb card
		blackWord := chosenWords[rand.Intn(len(chosenWords))]
		if card, ok := cards[blackWord]; ok {
			cards[blackWord] = db.Card{BelongsTo: "black", Guessed: false, Index: card.Index}
		}
		// Select red cards
		for j := 0; j < 8; j++ {
			randomWord := ""
			for {
				randomWord = chosenWords[rand.Intn(len(chosenWords))]
				if randomWord == blackWord {
					continue
				}
				if _, contains := utils.Contains(teamRedWords, randomWord); contains {
					continue
				}
				teamRedWords = append(teamRedWords, randomWord)
				break
			}
			if card, ok := cards[randomWord]; ok {
				cards[randomWord] = db.Card{BelongsTo: "red", Guessed: false, Index: card.Index}
			} else {
				log.Println("red not found", randomWord)
			}
		}
		// Select blue cards
		for j := 0; j < 9; j++ {
			randomWord := ""
			for {
				randomWord = chosenWords[rand.Intn(len(chosenWords))]
				if randomWord == blackWord {
					continue
				}
				if _, contains := utils.Contains(teamBlueWords, randomWord); contains {
					continue
				}
				if _, contains := utils.Contains(teamRedWords, randomWord); contains {
					continue
				}
				teamBlueWords = append(teamBlueWords, randomWord)
				break
			}
			if card, ok := cards[randomWord]; ok {
				cards[randomWord] = db.Card{BelongsTo: "blue", Guessed: false, Index: card.Index}
			} else {
				log.Println("blue not found", randomWord)
			}
		}
		db.UpdateGame(ctx, client, game.ID, map[string]interface{}{
			"status":          "running",
			"teamRed":         teamRed,
			"teamBlue":        teamBlue,
			"teamRedSpy":      teamRed[teamRedSpyID],
			"teamBlueSpy":     teamBlue[teamBlueSpyID],
			"teamRedGuesser":  teamRed[teamRedGuesserID],
			"teamBlueGuesser": teamBlue[teamBlueGuesserID],
			"cards":           cards,
			"whoseTurn":       "blue",
		})
	}
}

// HandlePlayerGuess takes in an action, determines if they player is allowed to make a guess, and processes the guess
func HandlePlayerGuess(ctx context.Context, client *firestore.Client, action string, playerID string, game *db.Game) {
	actionParts := strings.Split(action, " ")
	if len(actionParts) != 2 {
		log.Println("Received an incorrectly formatted guess", actionParts, playerID)
		return
	}
	word := actionParts[1]
	if playerCanGuess(game, playerID) {
		card, cardFound := game.Cards[word]
		if cardFound && !card.Guessed {
			newCards := map[string]db.Card{}
			for key, card := range game.Cards {
				newCards[key] = card
			}
			newCards[word] = db.Card{
				Index:     card.Index,
				BelongsTo: card.BelongsTo,
				Guessed:   true}
			status := game.Status
			whoseTurn := game.WhoseTurn
			if card.BelongsTo == "black" {
				whoseTurn = "over"
				if game.WhoseTurn == "red" {
					status = "bluewon"
				} else {
					status = "redwon"
				}
			} else if !playerGuessedCardCorrectly(game, &card, playerID) {
				if game.WhoseTurn == "red" {
					whoseTurn = "blue"
				} else {
					whoseTurn = "red"
				}
			} else {
				redCardsGuessed := 0
				blueCardsGuessed := 0
				for _, card := range newCards {
					if !card.Guessed {
						continue
					}
					if card.BelongsTo == "blue" {
						blueCardsGuessed++
					} else if card.BelongsTo == "red" {
						redCardsGuessed++
					}
				}
				if blueCardsGuessed == 9 {
					whoseTurn = "over"
					status = "bluewon"
				}
				if redCardsGuessed == 8 {
					whoseTurn = "over"
					status = "redwon"
				}
			}
			db.UpdateGame(ctx, client, game.ID, map[string]interface{}{
				"cards":                    newCards,
				"status":                   status,
				"whoseTurn":                whoseTurn,
				"lastCardGuessed":          word,
				"lastCardGuessedBy":        game.Players[playerID],
				"lastCardGuessedCorrectly": card.BelongsTo == game.WhoseTurn,
			})
		}
	}
}

// HandleEndTurn ends the turn for the given team
func HandleEndTurn(ctx context.Context, client *firestore.Client, game *db.Game, playerID string) {
	if playerCanEndTurn(game, playerID) {
		whoseTurn := game.WhoseTurn
		if game.WhoseTurn == "red" {
			whoseTurn = "blue"
		} else {
			whoseTurn = "red"
		}
		db.UpdateGame(ctx, client, game.ID, map[string]interface{}{
			"whoseTurn": whoseTurn,
		})
	}
}

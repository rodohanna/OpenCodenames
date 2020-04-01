package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"./game"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "/ %s!", r.Method)
}

// from the official documentation.
func firestoreExample() {
	ctx := context.Background()
	sa := option.WithCredentialsFile("./chunkynut-key.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	_, _, err = client.Collection("users").Add(ctx, map[string]interface{}{
		"first": "Ada",
		"last":  "Lovelace",
		"born":  1815,
	})
	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}
	_, _, err = client.Collection("users").Add(ctx, map[string]interface{}{
		"first":  "Alan",
		"middle": "Mathison",
		"last":   "Turing",
		"born":   1912,
	})
	if err != nil {
		log.Fatalf("Failed adding aturing: %v", err)
	}
	iter := client.Collection("users").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Println(doc.Data())
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	http.HandleFunc("/", index)
	http.HandleFunc("/game/create", game.CreateGame)
	http.HandleFunc("/game/join", game.JoinGame)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

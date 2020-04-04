package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"./game"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "/ %s!", r.Method)
}

func initFirestore() (*firestore.Client, error) {
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
	return client, nil
}

func main() {
	rand.Seed(time.Now().Unix())
	client, err := initFirestore()
	if err != nil {
		log.Fatalf("Failed initializing Firestore: %v", err)
	}
	defer client.Close()
	http.HandleFunc("/", index)
	http.HandleFunc("/game/create", game.CreateGameHandler(client))
	http.HandleFunc("/game/join", game.JoinGameHandler(client))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

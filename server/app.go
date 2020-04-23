package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/RobertDHanna/OpenCodenames/handlers"
	"github.com/RobertDHanna/OpenCodenames/hub"
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
	hub := hub.NewHub(client)
	go hub.Run()
	fs := http.FileServer(http.Dir("./static-assets"))
	http.Handle("/", fs)
	http.HandleFunc("/game/create", handlers.CreateGameHandler(client))
	http.HandleFunc("/game/join", handlers.JoinGameHandler(client))
	http.HandleFunc("/ws", handlers.PlayerHandler(client, hub))
	http.HandleFunc("/ws/spectate", handlers.SpectatorHandler(client, hub))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

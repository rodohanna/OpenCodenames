# OpenCodenames

A real-time implementation of Codenames created with React/TypeScript and Golang.

[You can play the game here!](https://www.chunkynut.com/#/)

## Installation

Stack:

- [React](https://reactjs.org/)
- [TypeScript](https://www.typescriptlang.org/)
- [Go](https://golang.org/doc/install)
- [Firestore](https://firebase.google.com/docs/firestore)

Requirements:

- [Yarn](https://classic.yarnpkg.com/en/)
- [Go CLI](https://golang.org/doc/install)
- [A Firestore database](https://firebase.google.com/docs/firestore)

Prerequisites:

- You will need a Google Firebase account. Create a Firestore database and place your application secret in the `server/` directory in a file named `chunkynut-key.json`
- If you desire [reCAPTCHA](https://developers.google.com/recaptcha/docs/v3) protection you will need to add a `recaptcha-key.txt` in the `server/` directory as well as replace the reCAPTCHA public keys in `client/public/index.html` and `client/src/hooks/useAPI.tsx`. If you don't want reCAPTCHA protection you will need to modify the code in `client/src/hooks/useAPI.tsx#executeRequest` and `server/handler/handlers.go#CreateGameHandler` to bypass the check.

Install dependencies and start the client

```bash
cd client
yarn && yarn start
```

Install dependencies and start the server

```bash
cd server
go mod download && go run app.go
```

## Architecture

The server hosts both the static assets for the client as well as the app code that provides the functionality.

### The Hub

Players join games and are placed in a "Hub" that maps games to "Clients". A Client is essentially just a WebSocket connection with additional data about the player (what their role is, is it their turn?, can they perform the action they just requested?, etc.). When an update happens to a game that one or more Clients are subscribed to, the Hub uses the Client's connection to broadcast the change.

### Firestore

Firestore allows the application to listen for real-time changes on a query/document/collection. A Goroutine is started when the app starts that listens for all changes on the "games" collection. When a change occurs, the Goroutine notifies the Hub of the change and Clients subscribed to the given game are notified.

### Browser

The browser app is essentially just a dumb client. It just receives a game from the server and displays it accordingly. All of the important logic happens on the server. A client is only ever given the information it needs for the particular role of any player. For example, a guesser doesn't receive the full state of the game and just filter the information out when displaying it. Only spies receive the full state of the game.

## Gallery

Coming Soon :)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[GPLv3](https://choosealicense.com/licenses/gpl-3.0/)

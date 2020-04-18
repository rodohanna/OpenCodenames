interface Message {
  Action: string;
}

type CardData = {
  BelongsTo: string;
  Guessed: boolean;
  Index: number;
};

type Game = {
  ID: string;
  Status: string;
  Players: string[];
  TeamRed: string[];
  TeamBlue: string[];
  Cards: { [x: string]: CardData };
  WhoseTurn: string;
  YourTurn: string;
  TeamRedSpy: string;
  TeamBlueSpy: string;
  TeamRedGuesser: string;
  TeamBlueGuesser: string;
  You: string;
};

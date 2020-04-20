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
  YourTurn: boolean;
  TeamRedSpy: string;
  TeamBlueSpy: string;
  TeamRedGuesser: string;
  TeamBlueGuesser: string;
  You: string;
  YouOwnGame: boolean;
  GameCanStart: boolean;
  LastCardGuessed: string;
  LastCardGuessedBy: string;
  LastCardGuessedCorrectly: boolean;
};

interface Toaster {
  blue: (message: string) => void;
  red: (message: string) => void;
  green: (message: string) => void;
  yellow: (message: string) => void;
}

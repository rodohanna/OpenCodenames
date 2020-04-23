interface Message {
  Action: string;
}

type CardData = {
  BelongsTo: string;
  Guessed: boolean;
  Index: number;
};

type BaseGame = {
  ID: string;
  Status: string;
  Players: string[];
  TeamRed: string[];
  TeamBlue: string[];
  TeamRedSpy: string;
  TeamBlueSpy: string;
  TeamRedGuesser: string;
  TeamBlueGuesser: string;
  WhoseTurn: string;
  LastCardGuessed: string;
  LastCardGuessedBy: string;
  LastCardGuessedCorrectly: boolean;
  Cards: { [x: string]: CardData };
};

type Game = {
  You: string;
  YouOwnGame: boolean;
  YourTurn: boolean;
  GameCanStart: boolean;
  BaseGame: BaseGame;
};

interface Toaster {
  blue: (message: string) => void;
  red: (message: string) => void;
  green: (message: string) => void;
  yellow: (message: string) => void;
}

declare var grecaptcha: any;

interface Message {
  Action: string;
}

type CardData = {
  BelongsTo: string;
  Guessed: boolean;
};

type Game = {
  ID: string;
  Status: string;
  Players: string[];
  TeamRed: string[];
  TeamBlue: string[];
  Cards: { [x: string]: CardData };
};

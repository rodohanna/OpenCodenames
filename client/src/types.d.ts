interface Message {
  Action: string;
}

type Game = {
  ID: string;
  Status: string;
  Players: string[];
  TeamRed: string[];
  TeamBlue: string[];
};

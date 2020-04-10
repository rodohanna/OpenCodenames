interface Message {
  body: string | null;
}

type Game = {
  ID: string;
  Status: string;
  Players: string[];
};

import React from 'react';
import Lobby from './Lobby';
import useQuery from './hooks/useQuery';
import useWebSocket from './hooks/useWebSocket';
import useLocalStorage from './hooks/useLocalStorage';
import { v4 as uuidv4 } from 'uuid';

function Game() {
  const query = useQuery();
  const gameID = query.get('gameID');
  const [game, setGame] = React.useState<Game | null>(null);
  // TODO: create usePlayerID hook
  const [playerID, setPlayerID] = useLocalStorage('playerID');
  const [connected, incomingMessage] = useWebSocket({
    webSocketUrl: `ws://localhost:8080/ws?gameID=${gameID}&playerID=${playerID}`,
    skip: typeof gameID !== 'string' && playerID !== null,
  });
  if (playerID === null) {
    setPlayerID(uuidv4());
  }
  React.useEffect(() => {
    setGame(incomingMessage);
  }, [incomingMessage]);
  if (typeof gameID !== 'string') {
    return <div>Invalid URL</div>;
  }
  if (game === null || !connected) {
    return <div>Loading</div>;
  }
  return <Lobby game={game} />;
}

export default Game;

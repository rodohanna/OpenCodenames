import React from 'react';
import Lobby from './Lobby';
import Board from './Board';
import useQuery from './hooks/useQuery';
import useWebSocket from './hooks/useWebSocket';
import usePlayerID from './hooks/userPlayerID';

function Game() {
  const query = useQuery();
  const isSpectator = query.has('spectate');
  const gameID = query.get('gameID');
  const [game, setGame] = React.useState<Game | null>(null);
  const playerID = usePlayerID();
  const webSocketHost = window.location.host.includes('localhost') ? 'localhost:8080' : window.location.host;
  const [connected, incomingMessage, sendMessage] = useWebSocket({
    webSocketUrl: isSpectator
      ? `ws://${webSocketHost}/ws/spectate?gameID=${gameID}`
      : `ws://${webSocketHost}/ws?gameID=${gameID}&playerID=${playerID}`,
    skip: typeof gameID !== 'string' && !isSpectator && playerID !== null,
  });
  React.useEffect(() => {
    setGame(incomingMessage);
  }, [incomingMessage]);
  if (typeof gameID !== 'string') {
    return <div>Invalid URL</div>;
  }
  if (game === null || !connected) {
    return <div>Loading</div>;
  }
  switch (game?.Status) {
    case 'pending': {
      return <Lobby game={game} sendMessage={sendMessage} />;
    }
    case 'running':
    case 'redwon':
    case 'bluewon': {
      return <Board game={game} sendMessage={sendMessage} />;
    }
    default: {
      return <div>Unknown Game State {JSON.stringify(game)}</div>;
    }
  }
}

export default Game;

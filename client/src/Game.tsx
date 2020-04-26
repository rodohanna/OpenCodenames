import React from 'react';
import Lobby from './Lobby';
import Board from './Board';
import useQuery from './hooks/useQuery';
import useWebSocket from './hooks/useWebSocket';
import { AppColor } from './config';
import { Loader, Message, Container } from 'semantic-ui-react';
import { v4 as uuidv4 } from 'uuid';
type GameProps = {
  appColor: AppColor;
  toaster: Toaster;
  setAppColor: (color: AppColor) => void;
};
function Game({ setAppColor, appColor, toaster }: GameProps) {
  const query = useQuery();
  const isSpectator = query.has('spectate');
  const gameID = query.get('gameID');
  const playerID = query.get('playerID');
  const [game, setGame] = React.useState<Game | null>(null);
  const [sessionID] = React.useState<string>(uuidv4());
  const webSocketHost = window.location.host.includes('localhost') ? 'localhost:8080' : window.location.host;
  const wsProtocol = window.location.protocol.includes('https') ? 'wss' : 'ws';
  const [connected, incomingMessage, sendMessage, reconnect] = useWebSocket({
    webSocketUrl: isSpectator
      ? `${wsProtocol}://${webSocketHost}/ws/spectate?gameID=${gameID}&sessionID=${sessionID}`
      : `${wsProtocol}://${webSocketHost}/ws?gameID=${gameID}&playerID=${playerID}&sessionID=${sessionID}`,
    skip: typeof gameID !== 'string' && !isSpectator && playerID !== null,
  });
  React.useEffect(() => {
    if (incomingMessage !== null) {
      setGame(incomingMessage);
    }
  }, [incomingMessage]);
  React.useEffect(() => {
    const intervalID = setInterval(() => {
      if (game && !connected) {
        reconnect();
      }
    }, 1000);
    return () => {
      clearInterval(intervalID);
    };
  }, [connected, reconnect, game]);
  if (typeof gameID !== 'string') {
    return (
      <Container>
        <Message negative>
          <Message.Header>Invalid URL</Message.Header>
          <p>Try rejoining or recreating the game.</p>
        </Message>
      </Container>
    );
  }
  if (game === null) {
    return <Loader size="massive" active />;
  }
  switch (game?.BaseGame?.Status) {
    case 'pending': {
      return <Lobby game={game} sendMessage={sendMessage} />;
    }
    case 'running':
    case 'redwon':
    case 'bluewon': {
      return (
        <Board game={game} sendMessage={sendMessage} appColor={appColor} setAppColor={setAppColor} toaster={toaster} />
      );
    }
    default: {
      return (
        <Container>
          <Message negative>
            <Message.Header>Unknown game state</Message.Header>
            <p>Please send the following to the developer:</p>
            <code style={{ wordWrap: 'break-word' }}>{JSON.stringify(game)}</code>
          </Message>
        </Container>
      );
    }
  }
}

export default Game;

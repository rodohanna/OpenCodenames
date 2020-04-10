import React from 'react';

type useWebSocketParams = {
  webSocketUrl: string;
  skip: boolean;
};

export default function ({
  webSocketUrl,
  skip,
}: useWebSocketParams): [boolean, Game | null, React.Dispatch<React.SetStateAction<Message>>] {
  const [socketUrl] = React.useState(webSocketUrl);
  const [socket, setSocket] = React.useState<WebSocket | null>(null);
  const [connected, setConnected] = React.useState(false);
  const [latestSentMessage, sendMessage] = React.useState<Message>({ body: null });
  const [incomingMessage, receiveMessage] = React.useState<Game | null>(null);
  React.useEffect(() => {
    if (!skip) {
      if (socket === null) {
        setSocket(new WebSocket(socketUrl));
      }
      socket?.addEventListener('open', (e) => {
        console.log('Opened  connection ', e);
        setConnected(true);
      });
      socket?.addEventListener('message', (e) => {
        console.log('Message from server ', e.data);
        receiveMessage(JSON.parse(e.data)?.game);
      });
      socket?.addEventListener('error', (e) => {
        console.error('WebSocket error ', e);
      });
      socket?.addEventListener('close', (e) => {
        console.log('Server closed connection ', e);
        setConnected(false);
      });
    }
  }, [socketUrl, socket, skip]);
  React.useEffect(() => {
    const preparedMessage = JSON.stringify({ Action: 'noop' });
    console.log('sending', preparedMessage);
    socket?.send(preparedMessage);
    // eslint-disable-next-line
  }, [latestSentMessage]);
  return [connected, incomingMessage, sendMessage];
}

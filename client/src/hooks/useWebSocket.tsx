import React from 'react';

type useWebSocketParams = {
  webSocketUrl: string;
  skip: boolean;
};

export default function ({
  webSocketUrl,
  skip,
}: useWebSocketParams): [boolean, Game | null, (message: string) => void] {
  const [socketUrl] = React.useState(webSocketUrl);
  const [socket, setSocket] = React.useState<WebSocket | null>(null);
  const [connected, setConnected] = React.useState(false);
  const [latestSentMessage, sendMessage] = React.useState<Message | null>(null);
  const [incomingMessage, receiveMessage] = React.useState<Game | null>(null);
  React.useEffect(() => {
    if (!skip) {
      if (socket === null) {
        setSocket(new WebSocket(socketUrl));
      }
      socket?.addEventListener('open', () => {
        setConnected(true);
      });
      socket?.addEventListener('message', (e) => {
        receiveMessage(JSON.parse(e.data)?.game);
      });
      socket?.addEventListener('error', (e) => {
        console.error('WebSocket error ', e);
      });
      socket?.addEventListener('close', () => {
        setConnected(false);
      });
    }
  }, [socketUrl, socket, skip]);
  React.useEffect(() => {
    const preparedMessage = JSON.stringify(latestSentMessage);
    console.log('sending', preparedMessage);
    socket?.send(preparedMessage);
    // eslint-disable-next-line
  }, [latestSentMessage]);
  const sendMessageWrapper = (message: string) => {
    sendMessage({ Action: message });
  };
  return [connected, incomingMessage, sendMessageWrapper];
}

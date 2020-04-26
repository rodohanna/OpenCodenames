import React from 'react';

type useWebSocketParams = {
  webSocketUrl: string;
  skip: boolean;
};

export default function ({
  webSocketUrl,
  skip,
}: useWebSocketParams): [boolean, Game | null, (message: string) => void, () => void] {
  const [socketUrl] = React.useState(webSocketUrl);
  const [socket, setSocket] = React.useState<WebSocket | null>(null);
  const [connected, setConnected] = React.useState(false);
  const [latestSentMessage, sendMessage] = React.useState<Message | null>(null);
  const [incomingMessage, receiveMessage] = React.useState<Game | null>(null);
  const [shouldReconnect, setShouldReconnect] = React.useState<boolean>(false);
  React.useEffect(() => {
    if (!skip) {
      if (shouldReconnect) {
        socket?.close();
        setSocket(null);
        setShouldReconnect(false);
        return;
      }
      if (socket === null) {
        setSocket(new WebSocket(socketUrl));
      }
      socket?.addEventListener('open', () => {
        setConnected(true);
      });
      socket?.addEventListener('message', (e) => {
        receiveMessage(JSON.parse(e.data));
      });
      socket?.addEventListener('error', (e) => {
        console.error('WebSocket error ', e);
      });
      socket?.addEventListener('close', () => {
        setConnected(false);
      });
    }
  }, [socketUrl, socket, skip, shouldReconnect]);
  React.useEffect(() => {
    const preparedMessage = JSON.stringify(latestSentMessage);
    console.log('sending', preparedMessage);
    socket?.send(preparedMessage);
    // eslint-disable-next-line
  }, [latestSentMessage]);
  const sendMessageWrapper = (message: string) => {
    sendMessage({ Action: message });
  };
  const reconnect = () => {
    console.log('attempting reconnect');
    setShouldReconnect(true);
  };
  return [connected, incomingMessage, sendMessageWrapper, reconnect];
}

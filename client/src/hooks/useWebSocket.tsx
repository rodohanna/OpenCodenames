import React from 'react';

export default function (): [boolean, Message, React.Dispatch<React.SetStateAction<Message>>] {
  const [socket, setSocket] = React.useState<WebSocket | null>(null);
  const [connected, setConnected] = React.useState(false);
  const [latestSentMessage, sendMessage] = React.useState<Message>({ body: null });
  const [incomingMessage, receiveMessage] = React.useState<Message>({ body: null });

  React.useEffect(() => {
    if (socket === null) {
      setSocket(new WebSocket('ws://localhost:8080/ws'));
    }
    socket?.addEventListener('open', (e) => {
      console.log('Opened  connection ', e);
      setConnected(true);
    });
    socket?.addEventListener('message', (e) => {
      console.log('Message from server ', e.data);
      receiveMessage({ body: e.data });
    });
    socket?.addEventListener('error', (e) => {
      console.error('WebSocket error ', e);
    });
    socket?.addEventListener('close', (e) => {
      console.log('Server closed connection ', e);
      setConnected(false);
    });
  }, [socket]);
  React.useEffect(() => {
    if (typeof latestSentMessage.body === 'string') {
      socket?.send(latestSentMessage.body);
    }
    // eslint-disable-next-line
  }, [latestSentMessage]);
  return [connected, incomingMessage, sendMessage];
}

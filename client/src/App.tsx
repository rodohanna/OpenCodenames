import React from 'react';
import './App.css';
import useWebSocket from './hooks/useWebSocket';

function App() {
  const inputRef = React.useRef<HTMLInputElement | null>(null);
  const [messages, setMessages] = React.useState<Array<string>>([]);
  const [connected, incomingMessage, sendMessage] = useWebSocket('ws://localhost:8080/ws?gameID=FAKE&playerID=FAKE2');
  React.useEffect(() => {
    if (typeof incomingMessage.body === 'string') {
      setMessages([...messages, incomingMessage.body]);
    }
    // eslint-disable-next-line
  }, [incomingMessage]);
  return (
    <div className="App">
      <label>Send:</label>
      <input
        ref={inputRef}
        type="text"
        id="send"
        name="send"
        onKeyDown={(e) => {
          if (e.keyCode === 13) {
            connected && sendMessage({ body: inputRef.current?.value || 'empty' });
          }
        }}
      />
      <br />
      <ul>
        {messages.map((message, index) => (
          <li key={`${message}${index}`}>{message}</li>
        ))}
      </ul>
    </div>
  );
}

export default App;

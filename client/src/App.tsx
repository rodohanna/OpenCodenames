import React from 'react';
import './App.css';
import 'semantic-ui-css/semantic.min.css';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import useWebSocket from './hooks/useWebSocket';
import Home from './Home';
import Lobby from './Lobby';

function App() {
  const [messages, setMessages] = React.useState<Array<string>>([]);
  const [connected, incomingMessage, sendMessage] = useWebSocket('ws://localhost:8080/ws?gameID=IKWE&playerID=abc123');
  console.log(connected, sendMessage);
  React.useEffect(() => {
    if (typeof incomingMessage.body === 'string') {
      setMessages([...messages, incomingMessage.body]);
    }
    // eslint-disable-next-line
  }, [incomingMessage]);
  return (
    <Router>
      <div className="App">
        <Switch>
          <Route path="/lobby">
            <Lobby />
          </Route>
          <Route path="/game">
            <div />
          </Route>
          <Route path="/">
            <Home />
          </Route>
        </Switch>
      </div>
    </Router>
  );
}

export default App;

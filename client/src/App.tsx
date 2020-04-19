import React from 'react';
import './App.css';
import 'semantic-ui-css/semantic.min.css';
import { HashRouter as Router, Switch, Route } from 'react-router-dom';
import Home from './Home';
import Game from './Game';

function App() {
  return (
    <Router>
      <div className="App">
        <Switch>
          <Route path="/game">
            <Game />
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

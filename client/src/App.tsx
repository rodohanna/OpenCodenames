import React from 'react';
import './App.css';
import 'semantic-ui-css/semantic.min.css';
import { HashRouter as Router, Switch, Route } from 'react-router-dom';
import Home from './Home';
import Game from './Game';
import { AppColor, AppColorToCSSColor } from './config';

function App() {
  const [appColor, setAppColor] = React.useState<AppColor>(AppColor.Blue);
  return (
    <Router>
      <div className="App" style={{ backgroundColor: AppColorToCSSColor[appColor] }}>
        <Switch>
          <Route path="/game">
            <Game appColor={appColor} setAppColor={(color: AppColor) => setAppColor(color)} />
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

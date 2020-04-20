import React from 'react';
import './App.css';
import 'semantic-ui-css/semantic.min.css';
import 'react-toastify/dist/ReactToastify.min.css';
import { HashRouter as Router, Switch, Route } from 'react-router-dom';
import Home from './Home';
import Game from './Game';
import { AppColor, AppColorToCSSColor } from './config';
import { ToastContainer, toast } from 'react-toastify';

toast.configure({
  autoClose: 3000,
});
const toaster: Toaster = {
  blue: (message: string) => toast.info(message),
  red: (message: string) => toast.error(message),
  green: (message: string) => toast.success(message),
};
function App() {
  const [appColor, setAppColor] = React.useState<AppColor>(AppColor.Blue);
  return (
    <Router>
      <div className="App" style={{ backgroundColor: AppColorToCSSColor[appColor] }}>
        <Switch>
          <Route path="/game">
            <Game appColor={appColor} setAppColor={(color: AppColor) => setAppColor(color)} toaster={toaster} />
          </Route>
          <Route path="/">
            <Home />
          </Route>
        </Switch>
        <ToastContainer />
      </div>
    </Router>
  );
}

export default App;

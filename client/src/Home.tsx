import React from 'react';
import { Container, Divider, Button, Form, Grid, Segment, Header, Icon, Checkbox, Popup } from 'semantic-ui-react';
import { useHistory } from 'react-router-dom';
import useAPI from './hooks/useAPI';
import useQuery from './hooks/useQuery';

function Home() {
  const query = useQuery();
  const history = useHistory();
  const [gameIDFieldRequiredError, setGameIDFieldRequiredError] = React.useState(false);
  const [playerNameFieldRequiredError, setPlayerNameFieldRequiredError] = React.useState(false);
  const [playingOnThisDevice, setPlayingOnThisDevice] = React.useState(true);
  const [joinGameID, setJoinGameID] = React.useState<string | null>(query.get('gameID'));
  const [playerName, setPlayerName] = React.useState<string | null>(null);
  const [shouldCreateGame, setShouldCreateGame] = React.useState(false);
  const [shouldJoinGame, setShouldJoinGame] = React.useState(false);
  const gameIDInParams = query.has('gameID');
  const [createGameLoading, createGameError, createGameResult] = useAPI({
    endpoint: '/game/create',
    method: 'POST',
    skip: !shouldCreateGame,
  });
  const [joinGameLoading, joinGameError, joinGameResult] = useAPI({
    endpoint: `/game/join?gameID=${joinGameID}&playerName=foo&playerID=bar`,
    method: 'POST',
    skip: !shouldJoinGame,
  });
  if (createGameResult?.id) {
    history.push(`/lobby?gameID=${createGameResult?.id}`);
  } else if (createGameError) {
    return <div>Something broke.. Try refreshing the page.</div>;
  }
  if (joinGameResult?.success) {
    history.push(`/lobby?gameID=${joinGameID}`);
  } else if (joinGameError) {
    return <div>Something broke.. Try refreshing the page.</div>;
  }
  console.log({ createGameLoading, joinGameLoading });
  return (
    <>
      <Container textAlign="center">
        <Header as="h2" icon inverted>
          <Icon name="gamepad" />
          OpenCodenames
          <Header.Subheader>Play the board game Codenames online with friends.</Header.Subheader>
        </Header>
      </Container>
      <Container textAlign="justified">
        <Divider />
        <Segment placeholder>
          <Grid columns={gameIDInParams ? 1 : 2} relaxed="very" stackable centered>
            <Grid.Column>
              <Form>
                <Form.Input
                  icon="add user"
                  iconPosition="left"
                  label="Enter a game ID"
                  placeholder="FRXX..."
                  value={joinGameID || ''}
                  onChange={(e) => {
                    if (e.target.value.length > 0 && gameIDFieldRequiredError) {
                      setGameIDFieldRequiredError(false);
                    }
                    setJoinGameID(e.target.value);
                  }}
                  error={gameIDFieldRequiredError}
                />
                <Form.Input
                  icon="add user"
                  iconPosition="left"
                  label="Enter a name"
                  placeholder="Morgana"
                  value={playerName || ''}
                  onChange={(e) => {
                    if (e.target.value.length > 0 && playerNameFieldRequiredError) {
                      setPlayerNameFieldRequiredError(false);
                    }
                    setPlayerName(e.target.value);
                  }}
                  error={playerNameFieldRequiredError}
                />
                <Button
                  content="Join game"
                  color="blue"
                  onClick={(_e) => {
                    const gameIDNotSet = joinGameID === null || joinGameID === '';
                    const playerNameNotSet = playerName === null || playerName === '';
                    if (gameIDNotSet) {
                      setGameIDFieldRequiredError(true);
                    }
                    if (playerNameNotSet) {
                      setPlayerNameFieldRequiredError(true);
                    }
                    if (gameIDNotSet || playerNameNotSet) {
                      return;
                    }
                    setShouldJoinGame(true);
                  }}
                />
              </Form>
            </Grid.Column>
            {!gameIDInParams && (
              <>
                <Divider vertical>Or</Divider>
                <Grid.Column verticalAlign="middle">
                  <div>
                    <Button
                      content="New game"
                      icon="add square"
                      size="big"
                      color="blue"
                      onClick={() => {
                        setShouldCreateGame(true);
                      }}
                    />
                    <br />
                    <Popup
                      content="Disabling this will require you to join the game on a different device."
                      trigger={
                        <Checkbox
                          label="I'll be playing on this device"
                          checked={playingOnThisDevice}
                          onChange={(_e) => {
                            setPlayingOnThisDevice(!playingOnThisDevice);
                          }}
                          toggle
                        />
                      }
                    />
                  </div>
                </Grid.Column>
              </>
            )}
          </Grid>
        </Segment>
      </Container>
    </>
  );
}

export default Home;

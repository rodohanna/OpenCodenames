import React from 'react';
import { Container, Divider, Button, Form, Grid, Segment, Header, Icon, Checkbox, Popup } from 'semantic-ui-react';
import { useHistory } from 'react-router-dom';
import useAPI from './hooks/useAPI';
import useQuery from './hooks/useQuery';
import useLocalStorage from './hooks/useLocalStorage';
import { v4 as uuidv4 } from 'uuid';

function Home() {
  const query = useQuery();
  const history = useHistory();
  const [gameIDFieldRequiredError, setGameIDFieldRequiredError] = React.useState(false);
  const [joinGamePlayerNameFieldRequiredError, setJoinGamePlayerNameFieldRequiredError] = React.useState(false);
  const [createGamePlayerNameFieldRequiredError, setCreateGamePlayerNameFieldRequiredError] = React.useState(false);
  const [playingOnThisDevice, setPlayingOnThisDevice] = React.useState(true);
  const [joinGameID, setJoinGameID] = React.useState<string | null>(query.get('gameID'));
  const [joinGamePlayerName, setJoinGamePlayerName] = React.useState<string | null>(null);
  const [createGamePlayerName, setCreateGamePlayerName] = React.useState<string | null>(null);
  const [shouldCreateGame, setShouldCreateGame] = React.useState(false);
  const [shouldJoinGame, setShouldJoinGame] = React.useState(false);
  const gameIDInParams = query.has('gameID');
  const [playerID, setPlayerID] = useLocalStorage('playerID');
  if (playerID === null) {
    setPlayerID(uuidv4());
  }
  const [createGameLoading, createGameError, createGameResult] = useAPI({
    endpoint: `/game/create${playingOnThisDevice ? `?playerID=${playerID}&playerName=${createGamePlayerName}` : ''}`,
    method: 'POST',
    skip: !shouldCreateGame || (playingOnThisDevice && (createGamePlayerName === null || createGamePlayerName === '')),
  });
  const [joinGameLoading, joinGameError, joinGameResult] = useAPI({
    endpoint: `/game/join?gameID=${joinGameID}&playerName=${joinGamePlayerName}&playerID=${playerID}`,
    method: 'POST',
    skip: !shouldJoinGame || joinGamePlayerName === null || joinGamePlayerName === '',
  });
  if (createGameResult?.id) {
    history.push(`/game?gameID=${createGameResult?.id}`);
  } else if (createGameError) {
    return <div>Something broke.. Try refreshing the page.</div>;
  }
  if (joinGameResult?.success) {
    history.push(`/game?gameID=${joinGameID}`);
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
                    setJoinGameID(e.target.value.replace(/\s/g, '').slice(0, 16).toLocaleUpperCase());
                  }}
                  error={gameIDFieldRequiredError}
                />
                <Form.Input
                  icon="add user"
                  iconPosition="left"
                  label="Enter a name"
                  placeholder="Morgana"
                  value={joinGamePlayerName || ''}
                  onChange={(e) => {
                    if (e.target.value.length > 0 && joinGamePlayerNameFieldRequiredError) {
                      setJoinGamePlayerNameFieldRequiredError(false);
                    }
                    setJoinGamePlayerName(e.target.value.replace(/\s/g, '').slice(0, 16));
                  }}
                  error={joinGamePlayerNameFieldRequiredError}
                />
                <Button
                  content="Join game"
                  color="blue"
                  onClick={(_e) => {
                    const gameIDNotSet = joinGameID === null || joinGameID === '';
                    const playerNameNotSet = joinGamePlayerName === null || joinGamePlayerName === '';
                    if (gameIDNotSet) {
                      setGameIDFieldRequiredError(true);
                    }
                    if (playerNameNotSet) {
                      setJoinGamePlayerNameFieldRequiredError(true);
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
                  <Form>
                    <Form.Input
                      icon="add user"
                      iconPosition="left"
                      label={<label style={{ textAlign: 'left' }}>Enter a name</label>}
                      placeholder="Ryuji"
                      value={createGamePlayerName || ''}
                      onChange={(e) => {
                        if (e.target.value.length > 0 && createGamePlayerNameFieldRequiredError) {
                          setCreateGamePlayerNameFieldRequiredError(false);
                        }
                        setCreateGamePlayerName(e.target.value.replace(/\s/g, '').slice(0, 16));
                      }}
                      error={createGamePlayerNameFieldRequiredError}
                      disabled={!playingOnThisDevice}
                    />
                    <Popup
                      content="Disabling this will require you to join the game on a different device."
                      trigger={
                        <Checkbox
                          style={{ marginBottom: '15px' }}
                          label="I'll be playing on this device"
                          checked={playingOnThisDevice}
                          onChange={(_e) => {
                            setPlayingOnThisDevice(!playingOnThisDevice);
                          }}
                          toggle
                        />
                      }
                    />
                    <Button
                      content="New game"
                      icon="add square"
                      size="big"
                      color="blue"
                      onClick={() => {
                        if (playingOnThisDevice && (createGamePlayerName === null || createGamePlayerName === '')) {
                          setCreateGamePlayerNameFieldRequiredError(true);
                          return;
                        }
                        setShouldCreateGame(true);
                      }}
                    />
                  </Form>
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

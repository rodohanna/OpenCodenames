import React from 'react';
import {
  Container,
  Divider,
  Button,
  Form,
  Grid,
  Segment,
  Header,
  Icon,
  Checkbox,
  Popup,
  Loader,
  Dimmer,
  Message,
} from 'semantic-ui-react';
import { useHistory } from 'react-router-dom';
import useAPI from './hooks/useAPI';
import useQuery from './hooks/useQuery';

function Home() {
  const query = useQuery();
  const history = useHistory();
  const [gameIDFieldRequiredError, setGameIDFieldRequiredError] = React.useState(false);
  const [joinGamePlayerNameFieldRequiredError, setJoinGamePlayerNameFieldRequiredError] = React.useState(false);
  const [joinGameGameError, setJoinGameGameError] = React.useState<any>(false);
  const [createGamePlayerNameFieldRequiredError, setCreateGamePlayerNameFieldRequiredError] = React.useState(false);
  const [playingOnThisDevice, setPlayingOnThisDevice] = React.useState(true);
  const [joinGameID, setJoinGameID] = React.useState<string | null>(query.get('gameID'));
  const [joinGamePlayerName, setJoinGamePlayerName] = React.useState<string | null>(null);
  const [createGamePlayerName, setCreateGamePlayerName] = React.useState<string | null>(null);
  const [shouldCreateGame, setShouldCreateGame] = React.useState(false);
  const [shouldJoinGame, setShouldJoinGame] = React.useState(false);
  const gameIDInParams = query.has('gameID');
  const [createGameLoading, createGameError, createGameResult] = useAPI({
    endpoint: `/game/create${playingOnThisDevice ? `?playerName=${createGamePlayerName}` : '?'}`,
    method: 'POST',
    skip: !shouldCreateGame || (playingOnThisDevice && (createGamePlayerName === null || createGamePlayerName === '')),
    withReCAPTCHA: true,
  });
  const [joinGameLoading, joinGameError, joinGameResult] = useAPI({
    endpoint: `/game/join?gameID=${joinGameID}&playerName=${joinGamePlayerName}`,
    method: 'POST',
    skip: !shouldJoinGame || joinGamePlayerName === null || joinGamePlayerName === '' || joinGameGameError,
    withReCAPTCHA: false,
  });
  React.useEffect(() => {
    if (joinGameResult?.error === 'GameDoesntExist') {
      setJoinGameGameError('The game could not be found');
    } else if (joinGameResult?.error === 'NameAlreadyTaken') {
      setJoinGameGameError('Someone in the game already has that name');
    } else if (joinGameResult?.error === 'GameIsFull') {
      setJoinGameGameError('That game is already full (8 players)');
    } else if (joinGameResult?.error === 'GameAlreadyStarted') {
      setJoinGameGameError('That game has already started');
    }
  }, [joinGameResult]);
  if (createGameResult?.id) {
    history.push(
      `/game?gameID=${createGameResult?.id}${
        !playingOnThisDevice ? '&spectate' : `&playerID=${createGameResult?.playerID}`
      }`,
    );
  } else if (createGameError) {
    return (
      <Container>
        <Message negative>
          <Message.Header>Something broke</Message.Header>
          <p>Try refreshing the page</p>
        </Message>
      </Container>
    );
  }
  if (joinGameResult?.success && joinGameResult?.playerID) {
    history.push(`/game?gameID=${joinGameID}&playerID=${joinGameResult?.playerID}`);
  } else if (joinGameError) {
    return (
      <Container>
        <Message negative>
          <Message.Header>Something broke</Message.Header>
          <p>Try refreshing the page</p>
        </Message>
      </Container>
    );
  }
  return (
    <>
      <Container textAlign="center">
        <Header as="h2" icon inverted>
          <Icon name="gamepad" />
          ChunkyNut
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
                  icon="puzzle"
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
                  error={gameIDFieldRequiredError || joinGameGameError}
                />
                <Form.Input
                  icon="add user"
                  iconPosition="left"
                  label="Enter your name"
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
                    setJoinGameGameError(false);
                  }}
                />
                <Dimmer active={joinGameLoading && !joinGameGameError}>
                  <Loader size="large">Loading</Loader>
                </Dimmer>
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
                      label={<label style={{ textAlign: 'left' }}>Enter your name</label>}
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
                    <Dimmer active={createGameLoading}>
                      <Loader size="large">Loading</Loader>
                    </Dimmer>
                    <div style={{ color: 'gray', fontSize: '7px', marginTop: '10px' }}>
                      <p style={{ margin: '0' }}>This site is protected by reCAPTCHA and the Google</p>
                      <a href="https://policies.google.com/privacy">Privacy Policy</a> and
                      <a href="https://policies.google.com/terms"> Terms of Service</a> apply.
                    </div>
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

import React from 'react';
import { Container, Divider, Button, Form, Grid, Segment, Header, Icon, Checkbox, Popup } from 'semantic-ui-react';
import { useHistory } from 'react-router-dom';

function Home() {
  const history = useHistory();
  const [fieldRequiredError, setFieldRequiredError] = React.useState(false);
  const [playingOnThisDevice, setPlayingOnThisDevice] = React.useState(true);
  const [joinGameID, setJoinGameID] = React.useState<string | null>(null);
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
          <Grid columns={2} relaxed="very" stackable centered>
            <Grid.Column>
              <Form>
                <Form.Input
                  icon="add user"
                  iconPosition="left"
                  label="Enter a game ID"
                  placeholder="FRXX..."
                  onChange={(e) => {
                    if (e.target.value.length > 0 && fieldRequiredError) {
                      setFieldRequiredError(false);
                    }
                    setJoinGameID(e.target.value);
                  }}
                  error={fieldRequiredError}
                />
                <Button
                  content="Join game"
                  color="blue"
                  onClick={(_e) => {
                    if (joinGameID === null || joinGameID === '') {
                      setFieldRequiredError(true);
                      return;
                    }
                    history.push('/lobby', { gameID: joinGameID });
                  }}
                />
              </Form>
            </Grid.Column>
            <Divider vertical>Or</Divider>
            <Grid.Column verticalAlign="middle">
              <div>
                <Button
                  content="New game"
                  icon="add square"
                  size="big"
                  color="blue"
                  onClick={() => {
                    history.push('/lobby', {
                      gameID: null,
                      willBePlayingOnThisDevice: playingOnThisDevice,
                    });
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
          </Grid>
        </Segment>
      </Container>
    </>
  );
}

export default Home;

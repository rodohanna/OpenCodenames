import React from 'react';
import { Container, Header, Icon, Divider, Card, Message, Button } from 'semantic-ui-react';

type LobbyProps = {
  game: Game;
  sendMessage: (message: string) => void;
};
function Lobby({ game, sendMessage }: LobbyProps) {
  const [startGameLoading, setStartGameLoading] = React.useState<boolean>(false);
  const joinLink = `${window.origin}/#/?gameID=${game.BaseGame.ID}`;
  const watchLink = `${window.origin}/#/game?gameID=${game.BaseGame.ID}&spectate`;
  return (
    <>
      <Container textAlign="center">
        <Header as="h2" icon inverted>
          <Icon name="stopwatch" />
          Lobby
          <Header.Subheader>Waiting on players</Header.Subheader>
        </Header>
      </Container>
      <Container textAlign="center">
        <Message
          compact
          header="Invite friends!"
          content={
            <>
              <div style={{ textAlign: 'left' }}>
                <span>
                  Join:{' '}
                  <a href={joinLink} target="_blank" rel="noopener noreferrer">
                    {joinLink}
                  </a>
                </span>
                <br />
                <span>
                  TV:{' '}
                  <a href={watchLink} target="_blank" rel="noopener noreferrer">
                    {watchLink}
                  </a>
                </span>
              </div>
              {game.YouOwnGame && (
                <>
                  <br />
                  <Button
                    onClick={() => {
                      sendMessage('StartGame');
                      setStartGameLoading(true);
                    }}
                    color="green"
                    disabled={!game.GameCanStart}
                    loading={startGameLoading}
                  >
                    Start game
                  </Button>
                </>
              )}
            </>
          }
        />
      </Container>
      <Container textAlign="justified">
        <Divider />
        <Card.Group centered>
          {game.BaseGame.Players.sort().map((playerName) => (
            <Card color="green" key={playerName}>
              <Card.Content>
                <Card.Description textAlign="center">
                  <Header as="h2" icon>
                    {playerName}
                  </Header>
                </Card.Description>
              </Card.Content>
            </Card>
          ))}
        </Card.Group>
      </Container>
    </>
  );
}

export default Lobby;

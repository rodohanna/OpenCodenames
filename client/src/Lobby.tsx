import React from 'react';
import { Container, Header, Icon, Divider, Card, Message, Button } from 'semantic-ui-react';

type LobbyProps = {
  game: Game;
  sendMessage: (message: string) => void;
};
function Lobby({ game, sendMessage }: LobbyProps) {
  const joinLink = `${window.origin}/#/?gameID=${game.ID}`;
  const watchLink = `${window.origin}/#/game?gameID=${game.ID}&spectate`;
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
            <div style={{ textAlign: 'left' }}>
              <span>
                Join: <a href={joinLink}>{joinLink}</a>
              </span>
              <br />
              <span>
                TV: <a href={watchLink}>{watchLink}</a>
              </span>
            </div>
          }
        />
      </Container>
      <Container>
        <Button onClick={() => sendMessage('StartGame')}>Start game</Button>
      </Container>
      <Container textAlign="justified">
        <Divider />
        <Card.Group centered>
          {game.Players.sort().map((playerName) => (
            <Card color="red" key={playerName}>
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

import React from 'react';
import { Container, Header, Icon, Divider, Card, Message } from 'semantic-ui-react';
import useQuery from './hooks/useQuery';

function Lobby() {
  const query = useQuery();
  if (!query.has('gameID')) {
    return <div>Invalid link.</div>;
  }
  const joinLink = `${window.origin}/join?gameID=${query.get('gameID')}`;
  const watchLink = `${window.origin}/tv?gameID=${query.get('gameID')}`;
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
      <Container textAlign="justified">
        <Divider />
        <Card.Group centered>
          <Card color="red">
            <Card.Content>
              <Card.Description textAlign="center">
                <Header as="h2" icon>
                  Chungo
                </Header>
              </Card.Description>
            </Card.Content>
          </Card>
        </Card.Group>
      </Container>
    </>
  );
}

export default Lobby;

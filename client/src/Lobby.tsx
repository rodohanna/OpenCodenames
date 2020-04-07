import React from 'react';
// import { useLocation } from 'react-router-dom';
import { Container, Header, Icon, Divider, Grid, SemanticCOLORS } from 'semantic-ui-react';

function Lobby() {
  //   const location = useLocation();
  const colors: Array<SemanticCOLORS> = [
    'red',
    'orange',
    'yellow',
    'olive',
    'green',
    'teal',
    'blue',
    'violet',
    'purple',
    'pink',
    'brown',
    'grey',
    'black',
  ];

  return (
    <>
      <Container textAlign="center">
        <Header as="h2" icon inverted>
          <Icon name="stopwatch" />
          Lobby
          <Header.Subheader>Waiting to start</Header.Subheader>
        </Header>
      </Container>
      <Container textAlign="justified">
        <Divider />

        <Grid columns={3} relaxed="very" padded>
          {colors.map((color) => (
            <Grid.Column color={color} key={color}>
              <Header as="h2" inverted textAlign="center">
                <Icon name="heart" size="tiny" />
                Name
              </Header>
            </Grid.Column>
          ))}
        </Grid>
      </Container>
    </>
  );
}

export default Lobby;

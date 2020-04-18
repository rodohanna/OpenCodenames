import React from 'react';
import { Divider, Container, Grid, Segment, List, Icon, Message } from 'semantic-ui-react';
import { chunk } from 'lodash';

type BoardProps = {
  game: Game;
  sendMessage: (message: string) => void;
};
function Board({ game, sendMessage }: BoardProps) {
  console.log(sendMessage);
  const gridRows = React.useMemo(() => {
    return chunk(
      Object.entries(game.Cards).sort((a, b) => {
        if (a[1].Index < b[1].Index) {
          return -1;
        } else if (a[1].Index > b[1].Index) {
          return 1;
        } else {
          return 0;
        }
      }),
      5,
    ).map((row) => {
      return (
        <Grid.Row>
          {row.map(([cardName, cardData]) => {
            return (
              <Grid.Column className="column-override">
                <Segment
                  className="game-segment"
                  textAlign="center"
                  style={{
                    userSelect: 'none',
                    ...(cardData.Guessed && { opacity: '.75' }),
                  }}
                  color={cardData.BelongsTo === 'red' ? 'red' : cardData.BelongsTo === 'blue' ? 'blue' : undefined}
                  inverted={['red', 'blue', 'black'].includes(cardData.BelongsTo)}
                >
                  {cardData.Guessed ? (
                    <div className="card-guessed">{cardName.toLocaleUpperCase()}</div>
                  ) : (
                    cardName.toLocaleUpperCase()
                  )}
                </Segment>
              </Grid.Column>
            );
          })}
        </Grid.Row>
      );
    });
  }, [game.Cards]);
  return (
    <Container textAlign="center">
      <Message size="big" color={game.YourTurn ? 'green' : game.WhoseTurn === 'red' ? 'red' : 'blue'}>
        {game.YourTurn ? 'Your Turn' : game.WhoseTurn === 'red' ? "Red's Turn" : "Blue's Turn"}
      </Message>
      <Segment padded>
        <Grid columns={2} textAlign="center">
          <Grid.Row>
            <Divider vertical fitted as="span">
              vs
            </Divider>
            <Grid.Column padded>
              <Icon name="chess knight" size="big" color="red" />
              <List verticalAlign="middle">
                {game.TeamRed.map((player) => (
                  <List.Item>
                    <List.Header style={{ color: player === game.You ? 'green' : 'black' }}>
                      {player}
                      {player === game.TeamRedSpy ? ' (spy)' : player === game.TeamRedGuesser ? ' (guesser)' : ''}
                    </List.Header>
                  </List.Item>
                ))}
              </List>
            </Grid.Column>
            <Grid.Column>
              <Icon name="chess bishop" size="big" color="blue" />
              <List verticalAlign="middle">
                {game.TeamBlue.map((player) => (
                  <List.Item>
                    <List.Header style={{ color: player === game.You ? 'green' : 'black' }}>
                      {player}
                      {player === game.TeamBlueSpy ? ' (spy)' : player === game.TeamBlueGuesser ? ' (guesser)' : ''}
                    </List.Header>
                  </List.Item>
                ))}
              </List>
            </Grid.Column>
          </Grid.Row>
        </Grid>
      </Segment>
      <Grid stackable columns={5} container celled="internally" style={{ backgroundColor: 'cornflowerblue' }}>
        {gridRows}
      </Grid>
    </Container>
  );
}

export default Board;

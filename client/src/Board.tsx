import React from 'react';
import { Grid, Segment } from 'semantic-ui-react';
import { chunk } from 'lodash';

type BoardProps = {
  game: Game;
  sendMessage: (message: string) => void;
};
function Board({ game, sendMessage }: BoardProps) {
  console.log(sendMessage);
  const gridRows = chunk(Object.entries(game.Cards), 5).map((row) => {
    return (
      <Grid.Row>
        {row.map(([cardName]) => {
          return (
            <Grid.Column>
              <Segment>{cardName}</Segment>
            </Grid.Column>
          );
        })}
      </Grid.Row>
    );
  });
  return <Grid columns="equal">{gridRows}</Grid>;
}

export default Board;

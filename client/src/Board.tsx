import React from 'react';
import { Grid, Segment } from 'semantic-ui-react';
import { chunk } from 'lodash';

type BoardProps = {
  game: Game;
  sendMessage: (message: string) => void;
};
function Board({ game, sendMessage }: BoardProps) {
  console.log(sendMessage);
  const gridRows = React.useMemo(() => {
    return chunk(Object.entries(game.Cards), 5).map((row) => {
      return (
        <Grid.Row>
          {row.map(([cardName]) => {
            return (
              <Grid.Column className="column-override">
                <Segment textAlign="center" style={{ userSelect: 'none' }}>
                  {cardName.toLocaleUpperCase()}
                </Segment>
              </Grid.Column>
            );
          })}
        </Grid.Row>
      );
    });
  }, [game.Cards]);
  return (
    <Grid stackable columns={5} container celled="internally" style={{ backgroundColor: 'cornflowerblue' }}>
      {gridRows}
    </Grid>
  );
}

export default Board;

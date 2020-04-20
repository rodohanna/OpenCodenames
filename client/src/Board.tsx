import React from 'react';
import { Divider, Container, Grid, Segment, List, Icon, Message, Button, Loader } from 'semantic-ui-react';
import { chunk } from 'lodash';
import { AppColor, AppColorToCSSColor } from './config';

type BoardProps = {
  game: Game;
  appColor: AppColor;
  toaster: Toaster;
  sendMessage: (message: string) => void;
  setAppColor: (color: AppColor) => void;
};
type BannerMessageProps = {
  game: Game;
};
function BannerMessage({ game }: BannerMessageProps) {
  if (game.Status === 'redwon') {
    return (
      <Message size="big" color={game.TeamRed.includes(game.You) ? 'green' : 'yellow'}>
        Red Team Won!
      </Message>
    );
  } else if (game.Status === 'bluewon') {
    return (
      <Message size="big" color={game.TeamBlue.includes(game.You) ? 'green' : 'yellow'}>
        Blue Team Won!
      </Message>
    );
  }
  return (
    <Message size="big" color={game.YourTurn ? 'green' : game.WhoseTurn === 'red' ? 'red' : 'blue'}>
      {game.YourTurn ? 'Your Turn' : game.WhoseTurn === 'red' ? "Red's Turn" : "Blue's Turn"}
    </Message>
  );
}
function Board({ game, sendMessage, appColor, setAppColor, toaster }: BoardProps) {
  const gameIsRunning = game.Status === 'running';
  const playerIsOnTeamRed = game.TeamRed.includes(game.You);
  const playerIsOnTeamBlue = game.TeamBlue.includes(game.You);
  const isPlayersTurn =
    (playerIsOnTeamRed && game.WhoseTurn === 'red') || (playerIsOnTeamBlue && game.WhoseTurn === 'blue');
  const playerIsGuesser = game.TeamRedGuesser === game.You || game.TeamBlueGuesser === game.You;
  const [loadingWord, setLoadingWord] = React.useState<string | null>(null);
  React.useEffect(() => {
    setLoadingWord(null);
  }, [game.Cards]);
  React.useEffect(() => {
    if (playerIsOnTeamRed) {
      setAppColor(AppColor.Red);
    } else if (playerIsOnTeamBlue) {
      setAppColor(AppColor.Blue);
    }
  }, [playerIsOnTeamRed, playerIsOnTeamBlue, setAppColor]);
  React.useEffect(() => {
    if (isPlayersTurn) {
      toaster.green("🎉 It's your team's turn!");
    } else if (game.WhoseTurn === 'blue') {
      toaster.blue("👿 It's the Blue team's turn");
    } else if (game.WhoseTurn === 'red') {
      toaster.red("👿 It's the Red team's turn");
    }
  }, [game.WhoseTurn, isPlayersTurn, toaster]);
  React.useEffect(() => {
    if (game.LastCardGuessed !== '' && game.LastCardGuessedBy !== '') {
      if (game.LastCardGuessedCorrectly) {
        toaster.green(`😊 ${game.LastCardGuessedBy} guessed "${game.LastCardGuessed.toLocaleUpperCase()}" correctly`);
      } else {
        toaster.yellow(
          `😞 ${game.LastCardGuessedBy} guessed "${game.LastCardGuessed.toLocaleUpperCase()}" incorrectly`,
        );
      }
    }
  }, [game.LastCardGuessed, game.LastCardGuessedBy, game.LastCardGuessedCorrectly, toaster]);
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
    ).map((row, index) => {
      return (
        <Grid.Row key={index}>
          {row.map(([cardName, cardData]) => {
            return (
              <Grid.Column key={cardName} className="column-override">
                <Segment
                  className={gameIsRunning ? 'game-segment' : ''}
                  textAlign="center"
                  style={{
                    userSelect: 'none',
                    ...((cardData.Guessed || !gameIsRunning) && { opacity: '.75' }),
                  }}
                  color={cardData.BelongsTo === 'red' ? 'red' : cardData.BelongsTo === 'blue' ? 'blue' : undefined}
                  inverted={['red', 'blue', 'black'].includes(cardData.BelongsTo)}
                  onClick={() => {
                    if ([game.TeamBlueGuesser, game.TeamRedGuesser].includes(game.You) && game.YourTurn) {
                      sendMessage(`Guess ${cardName}`);
                      setLoadingWord(cardName);
                    }
                  }}
                  disabled={!gameIsRunning}
                >
                  {cardName === loadingWord ? (
                    <Loader active inline size="tiny" />
                  ) : cardData.Guessed ? (
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
  }, [
    game.Cards,
    game.TeamBlueGuesser,
    game.TeamRedGuesser,
    game.You,
    game.YourTurn,
    sendMessage,
    gameIsRunning,
    loadingWord,
  ]);
  return (
    <Container textAlign="center">
      <BannerMessage game={game} />
      <Segment padded>
        <Grid columns={2} textAlign="center">
          <Grid.Row>
            <Divider vertical fitted as="span">
              vs
            </Divider>
            <Grid.Column padded="true">
              <Icon name="chess knight" size="big" color="red" />
              <List verticalAlign="middle">
                {game.TeamRed.sort().map((player) => (
                  <List.Item key={player}>
                    <List.Header style={{ color: player === game.You ? 'green' : 'black' }}>
                      {player}
                      {player === game.TeamRedSpy ? ' (spy)' : player === game.TeamRedGuesser ? ' (guesser)' : ''}
                    </List.Header>
                  </List.Item>
                ))}
              </List>
              {playerIsOnTeamRed && playerIsGuesser && (
                <Button color="red" disabled={!game.YourTurn} onClick={() => sendMessage('EndTurn')}>
                  End Turn
                </Button>
              )}
            </Grid.Column>
            <Grid.Column>
              <Icon name="chess bishop" size="big" color="blue" />
              <List verticalAlign="middle">
                {game.TeamBlue.sort().map((player) => (
                  <List.Item key={player}>
                    <List.Header style={{ color: player === game.You ? 'green' : 'black' }}>
                      {player}
                      {player === game.TeamBlueSpy ? ' (spy)' : player === game.TeamBlueGuesser ? ' (guesser)' : ''}
                    </List.Header>
                  </List.Item>
                ))}
              </List>
              {playerIsOnTeamBlue && playerIsGuesser && (
                <Button color="red" disabled={!game.YourTurn} onClick={() => sendMessage('EndTurn')}>
                  End Turn
                </Button>
              )}
            </Grid.Column>
          </Grid.Row>
        </Grid>
      </Segment>
      <Grid
        stackable
        columns={5}
        container
        celled="internally"
        style={{ backgroundColor: AppColorToCSSColor[appColor] }}
      >
        {gridRows}
      </Grid>
    </Container>
  );
}

export default Board;

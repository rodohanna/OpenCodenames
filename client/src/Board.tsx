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
  sendMessage: (message: string) => void;
};
function BannerMessage({ game, sendMessage }: BannerMessageProps) {
  const { You } = game;
  const [restartingGame, setRestartingGame] = React.useState(false);
  const _BannerMessage = function (
    message: string,
    color: 'green' | 'yellow' | 'red' | 'blue',
    startNewGame: boolean,
    sendMessage: (message: string) => void,
  ) {
    return (
      <Message size="big" color={color}>
        {message}
        {startNewGame && (
          <>
            <br />
            <Button
              color="green"
              onClick={() => {
                setRestartingGame(true);
                sendMessage('RestartGame');
              }}
              disabled={!game.YouOwnGame || restartingGame}
              loading={restartingGame}
            >
              Restart game
            </Button>
          </>
        )}
      </Message>
    );
  };
  const {
    YourTurn,
    BaseGame: { Status, TeamRed, TeamBlue, WhoseTurn },
  } = game;
  if (Status === 'redwon') {
    return _BannerMessage('Red Team Won!', TeamRed.includes(You) ? 'green' : 'yellow', true, sendMessage);
  } else if (Status === 'bluewon') {
    return _BannerMessage('Blue Team Won!', TeamBlue.includes(You) ? 'green' : 'yellow', true, sendMessage);
  }
  return _BannerMessage(
    YourTurn ? 'Your Turn' : WhoseTurn === 'red' ? "Red's Turn" : "Blue's Turn",
    YourTurn ? 'green' : WhoseTurn === 'red' ? 'red' : 'blue',
    false,
    sendMessage,
  );
}

function TeamDescription({
  icon,
  color,
  team,
  you,
  spy,
  guesser,
  yourTurn,
  endTurnLoading,
  setEndTurnLoading,
  sendMessage,
}: {
  icon: 'chess knight' | 'chess bishop';
  color: 'red' | 'blue';
  team: string[];
  you: string;
  spy: string;
  guesser: string;
  yourTurn: boolean;
  endTurnLoading: boolean;
  setEndTurnLoading: (isLoading: boolean) => void;
  sendMessage: (message: string) => void;
}) {
  const youAreGuesser = you === guesser;
  return (
    <>
      <Icon name={icon} size="big" color={color} />
      <List verticalAlign="middle">
        {team.sort().map((player) => (
          <List.Item key={player}>
            <List.Header style={{ color: player === you ? 'green' : 'black' }}>
              {player}
              {player === spy ? ' (spy)' : player === guesser ? ' (guesser)' : ''}
            </List.Header>
          </List.Item>
        ))}
      </List>
      {youAreGuesser && (
        <Button
          color="red"
          disabled={!yourTurn}
          onClick={() => {
            sendMessage('EndTurn');
            setEndTurnLoading(true);
          }}
          loading={endTurnLoading}
          negative
        >
          End Turn
        </Button>
      )}
    </>
  );
}
function Board({ game, sendMessage, appColor, setAppColor, toaster }: BoardProps) {
  const {
    You,
    YourTurn,
    BaseGame: {
      Cards,
      Status,
      TeamRed,
      TeamBlue,
      WhoseTurn,
      TeamBlueGuesser,
      TeamRedGuesser,
      LastCardGuessed,
      LastCardGuessedBy,
      LastCardGuessedCorrectly,
      TeamRedSpy,
      TeamBlueSpy,
    },
  } = game;
  const gameIsRunning = Status === 'running';
  const playerIsOnTeamRed = TeamRed.includes(You);
  const playerIsOnTeamBlue = TeamBlue.includes(You);
  const isPlayersTurn = (playerIsOnTeamRed && WhoseTurn === 'red') || (playerIsOnTeamBlue && WhoseTurn === 'blue');
  const [loadingWord, setLoadingWord] = React.useState<string | null>(null);
  const [endTurnLoading, setEndTurnLoading] = React.useState<boolean>(false);
  if (endTurnLoading && !isPlayersTurn) {
    setEndTurnLoading(false);
  }
  React.useEffect(() => {
    setLoadingWord(null);
  }, [Cards]);
  React.useEffect(() => {
    if ((Status === 'redwon' && playerIsOnTeamRed) || (Status === 'bluewon' && playerIsOnTeamBlue)) {
      toaster.green('Your team won!');
    } else if (['redwon', 'bluewon'].includes(Status)) {
      toaster.yellow('Your team lost');
    }
  }, [Status, playerIsOnTeamRed, playerIsOnTeamBlue, toaster]);
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
    } else if (WhoseTurn === 'blue') {
      toaster.blue("👿 It's the Blue team's turn");
    } else if (WhoseTurn === 'red') {
      toaster.red("👿 It's the Red team's turn");
    }
  }, [WhoseTurn, isPlayersTurn, toaster]);
  React.useEffect(() => {
    if (LastCardGuessed !== '' && LastCardGuessedBy !== '') {
      if (LastCardGuessedCorrectly) {
        toaster.green(`😊 ${LastCardGuessedBy} guessed "${LastCardGuessed.toLocaleUpperCase()}" correctly`);
      } else {
        toaster.yellow(`😞 ${LastCardGuessedBy} guessed "${LastCardGuessed.toLocaleUpperCase()}" incorrectly`);
      }
    }
  }, [LastCardGuessed, LastCardGuessedBy, LastCardGuessedCorrectly, toaster]);
  const gridRows = React.useMemo(() => {
    return chunk(
      Object.entries(Cards).sort((a, b) => {
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
                    if ([TeamBlueGuesser, TeamRedGuesser].includes(You) && YourTurn) {
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
  }, [Cards, TeamBlueGuesser, TeamRedGuesser, You, YourTurn, sendMessage, gameIsRunning, loadingWord]);
  return (
    <Container textAlign="center">
      <BannerMessage game={game} sendMessage={sendMessage} />
      <Segment padded>
        <Grid columns={2} textAlign="center">
          <Grid.Row>
            <Divider vertical fitted as="span">
              vs
            </Divider>
            <Grid.Column padded="true">
              <TeamDescription
                icon="chess knight"
                color="red"
                team={TeamRed}
                you={You}
                spy={TeamRedSpy}
                guesser={TeamRedGuesser}
                yourTurn={YourTurn}
                sendMessage={sendMessage}
                endTurnLoading={endTurnLoading}
                setEndTurnLoading={setEndTurnLoading}
              />
            </Grid.Column>
            <Grid.Column>
              <TeamDescription
                icon="chess bishop"
                color="blue"
                team={TeamBlue}
                you={You}
                spy={TeamBlueSpy}
                guesser={TeamBlueGuesser}
                yourTurn={YourTurn}
                sendMessage={sendMessage}
                endTurnLoading={endTurnLoading}
                setEndTurnLoading={setEndTurnLoading}
              />
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

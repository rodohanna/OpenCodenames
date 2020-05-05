import React from 'react';
import { Divider, Container, Grid, Segment, List, Icon, Message, Button, Loader } from 'semantic-ui-react';
import { chunk } from 'lodash';
import { AppColor, AppColorToCSSColor } from './config';
import useLocalStorage from './hooks/useLocalStorage';

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
              disabled={restartingGame}
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
  cardsLeft,
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
  cardsLeft: number;
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
  const cardsLeftText = cardsLeft !== 1 ? `cards left` : `card left`;
  return (
    <>
      <Icon name={icon} size="big" color={color} />
      {
        <Container style={{ marginTop: '5px' }}>
          <span style={{ fontSize: '17px' }}>
            <i>
              <b>{cardsLeft}</b> {cardsLeftText}
            </i>
          </span>
        </Container>
      }
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
  const [hasSeenTutorial, setHasSeenTutorialRerender, setHasSeenTutorialNoRerender] = useLocalStorage(
    'has-seen-tutorial',
    'false',
  );
  const gameIsRunning = Status === 'running';
  const playerIsOnTeamRed = TeamRed.includes(You);
  const playerIsOnTeamBlue = TeamBlue.includes(You);
  const isPlayersTurn = (playerIsOnTeamRed && WhoseTurn === 'red') || (playerIsOnTeamBlue && WhoseTurn === 'blue');
  const [loadingWord, setLoadingWord] = React.useState<string | null>(null);
  const [endTurnLoading, setEndTurnLoading] = React.useState<boolean>(false);
  const { blueCardsLeft, redCardsLeft } = Object.values(game.BaseGame.Cards).reduce(
    (accum, card) => {
      return {
        ...accum,
        ...(card.BelongsTo === 'blue' && card.Guessed && { blueCardsLeft: accum.blueCardsLeft - 1 }),
        ...(card.BelongsTo === 'red' && card.Guessed && { redCardsLeft: accum.redCardsLeft - 1 }),
      };
    },
    // TODO: These are magic numbers right now, I should make them configurable later
    { blueCardsLeft: 9, redCardsLeft: 8 },
  );
  React.useEffect(() => {
    if (hasSeenTutorial === 'false') {
      setHasSeenTutorialNoRerender('true');
    }
  }, [hasSeenTutorial, setHasSeenTutorialNoRerender]);
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
                  className="game-segment"
                  textAlign="center"
                  style={{
                    userSelect: 'none',
                    ...((cardData.Guessed || !gameIsRunning) && { opacity: '.75' }),
                  }}
                  color={cardData.BelongsTo === 'red' ? 'red' : cardData.BelongsTo === 'blue' ? 'blue' : undefined}
                  inverted={['red', 'blue', 'black'].includes(cardData.BelongsTo)}
                  onClick={() => {
                    if (
                      [TeamBlueGuesser, TeamRedGuesser].includes(You) &&
                      YourTurn &&
                      loadingWord === null &&
                      !cardData.Guessed
                    ) {
                      sendMessage(`Guess ${cardName}`);
                      setLoadingWord(cardName);
                    }
                  }}
                  disabled={!gameIsRunning}
                >
                  <div>
                    {cardName === loadingWord ? (
                      <Loader active inline size="tiny" />
                    ) : cardData.Guessed ? (
                      <div className="card-guessed">{cardName.toLocaleUpperCase()}</div>
                    ) : (
                      cardName.toLocaleUpperCase()
                    )}
                  </div>
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
      {hasSeenTutorial === 'false' && (
        <Message onDismiss={() => setHasSeenTutorialRerender('true')} floating info size="large">
          <Message.Header>How To Play</Message.Header>
          <p>
            After the Spy gives a clue the Guesser needs to <b>click on a card</b> in order to guess.
            <br />
            <br />
            The card's color will be revealed and if the card was guessed correctly, the Guesser will have the option to{' '}
            <b>continue guessing</b> or <b>end their turn</b>.
            <br />
            <br />
            If the Guesser guesses incorrectly their turn will <b>automatically end</b>.
            <br />
            <br />
            Remember, the Guesser must click on the <b>End Turn</b> button if they choose to stop guessing.
            <br />
            <br />
            Have fun!
          </p>
        </Message>
      )}
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
                cardsLeft={redCardsLeft}
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
                cardsLeft={blueCardsLeft}
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
      <Grid columns={5} celled="internally" style={{ backgroundColor: AppColorToCSSColor[appColor] }}>
        {gridRows}
      </Grid>
    </Container>
  );
}

export default Board;

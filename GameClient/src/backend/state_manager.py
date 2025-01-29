# Now you can import the package
import logging
import threading

from statemachine import StateMachine, State


class GameStateMachine(StateMachine):
    # Initialize the lock
    _lock = threading.Lock()

    # Define states
    stateForkMyTurn = State('ForkMyTurn')
    stateGame = State('game')
    stateLobby = State('Lobby')
    stateMyTurn = State('MyTurn')
    stateNextDice = State('NextDice')
    stateRunningGame = State('Running_Game')
    stateStart = State('Start', initial=True)
    stateForkNextDice = State('ForkNextDice')
    stateReconnect = State('Reconnect')


    # Define transitions
    ClientLogin = (
        stateStart.to(stateLobby)
    )

    # region Reconnect
    ClientReconnect = (
        stateStart.to(stateReconnect)
    )
    ResponseServerReconnectBeforeGame = (
        stateReconnect.to(stateGame)
    )
    ResponseServerReconnectRunningGame = (
        stateReconnect.to(stateRunningGame)
    )
    ResponseServerGameList = (
        stateReconnect.to(stateLobby)
    )
    # endregion

    ServerUpdateGameList = stateLobby.to(stateLobby)
    ClientJoinGame = stateLobby.to(stateGame)
    ClientCreateGame = stateLobby.to(stateGame)

    ClientStartGame = stateGame.to(stateRunningGame)
    ServerUpdateStartGame = (
        stateGame.to(stateRunningGame)
    )
    ServerUpdatePlayerList = stateGame.to(stateGame)


    ServerUpdateEndScore = (stateRunningGame.to(stateLobby)
                            )
    ServerUpdateNotEnoughPlayers = (
        stateRunningGame.to(stateLobby)
    )

    ServerStartTurn = stateRunningGame.to(stateMyTurn)
    ServerUpdateGameData = (stateRunningGame.to(stateRunningGame)
                            | stateMyTurn.to(stateMyTurn)
                            | stateNextDice.to(stateNextDice)
                            )

    ClientRollDice = stateMyTurn.to(stateForkMyTurn)
    ServerPingPlayer = (
            stateMyTurn.to(stateMyTurn) |
            stateNextDice.to(stateNextDice) |
            stateGame.to(stateGame) |
            stateRunningGame.to(stateRunningGame) |
            stateLobby.to(stateLobby)
    )

    ResponseServerDiceEndTurn = stateForkMyTurn.to(stateRunningGame)
    ResponseServerDiceNext = stateForkMyTurn.to(stateNextDice)

    ClientEndTurn = stateNextDice.to(stateRunningGame)
    ClientNextDice = stateNextDice.to(stateForkNextDice)

    ResponseServerNextDiceEndScore = stateForkNextDice.to(stateLobby)
    ResponseServerNextDiceSuccess = stateForkNextDice.to(stateMyTurn)

    def can_fire(self, trigger: str) -> bool:
        with self._lock:
            return any(event.id == trigger for event in self.allowed_events)

    def send_trigger(self, trigger: str):
        with self._lock:
            before = self.current_state_value
            self.send(trigger)
            after = self.current_state_value
            logging.info(f"AUTOMATA: State transition: {before} -> {trigger} -> {after}")

    def get_current_state(self):
        with self._lock:
            return self.current_state_value

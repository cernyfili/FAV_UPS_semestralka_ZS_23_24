# Now you can import the package
import threading

from statemachine import StateMachine, State


class GameStateMachine(StateMachine):
    # Initialize the lock
    _lock = threading.Lock()

    # Define states
    stateEnd = State('End', final=True)
    stateErrorGame = State('ErrorGame')
    stateErrorLobby = State('ErrorLobby')
    stateErrorRunningGame = State('ErrorRunning_Game')
    stateForkMyTurn = State('ForkMyTurn')
    stateGame = State('game')
    stateLobby = State('Lobby')
    stateMyTurn = State('MyTurn')
    stateNextDice = State('NextDice')
    stateRunningGame = State('Running_Game')
    stateStart = State('Start', initial=True)
    stateForkNextDice = State('ForkNextDice')

#todo check if this is correct
    # Define transitions
    ClientLogin = stateStart.to(stateLobby)
    ClientLogout = stateLobby.to(stateEnd)
    ServerUpdateGameList = stateLobby.to(stateLobby)
    ClientJoinGame = stateLobby.to(stateGame)
    ClientCreateGame = stateLobby.to(stateGame)

    ServerReconnectGameList = stateErrorLobby.to(stateLobby)


    ClientStartGame = stateGame.to(stateRunningGame)
    ServerUpdateStartGame = (stateGame.to(stateRunningGame)
                             | stateErrorGame.to(stateErrorRunningGame)
                             )
    ServerUpdatePlayerList = stateGame.to(stateGame)
    ErrorPlayerUnreachable = (stateGame.to(stateErrorGame)
                              | stateMyTurn.to(stateErrorRunningGame)
                              | stateRunningGame.to(stateErrorRunningGame)
                              | stateLobby.to(stateErrorLobby)
                              )
    ServerReconnectPlayerList = stateErrorGame.to(stateGame)

    ServerUpdateEndScore = (stateRunningGame.to(stateLobby)
                            | stateErrorRunningGame.to(stateErrorGame))
    ServerStartTurn = stateRunningGame.to(stateMyTurn)
    ServerUpdateGameData = stateRunningGame.to(stateRunningGame)

    ServerReconnectGameData = stateErrorRunningGame.to(stateRunningGame)

    ClientRollDice = stateMyTurn.to(stateForkMyTurn)
    ServerPingPlayer = (stateMyTurn.to(stateMyTurn)
                        | stateNextDice.to(stateNextDice))

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
            self.send(trigger)

    def get_current_state(self):
        with self._lock:
            return self.current_state_value


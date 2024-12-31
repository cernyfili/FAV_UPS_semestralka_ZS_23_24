from statemachine import StateMachine, State


class GameStateMachine(StateMachine):
    # Define states
    stateEnd = State('End')
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
    ErrorPlayerUnreachable = stateLobby.to(stateErrorLobby)
    ServerReconnectGameList = stateErrorLobby.to(stateLobby)
    ClientStartGame = stateGame.to(stateRunningGame)
    ServerUpdateStartGame = stateGame.to(stateRunningGame)
    ServerUpdatePlayerList = stateGame.to(stateGame)
    ErrorPlayerUnreachable = (stateGame.to(stateErrorGame)
                              | stateMyTurn.to(stateErrorRunningGame)
                              | stateRunningGame.to(stateErrorRunningGame)
                              )
    ServerReconnectPlayerList = stateErrorGame.to(stateGame)
    ServerUpdateStartGame = stateErrorGame.to(stateErrorRunningGame)
    ServerUpdateEndScore = stateRunningGame.to(stateLobby)
    ServerStartTurn = stateRunningGame.to(stateMyTurn)
    ServerUpdateGameData = stateRunningGame.to(stateRunningGame)

    ServerReconnectGameData = stateErrorRunningGame.to(stateRunningGame)
    ServerUpdateEndScore = stateErrorRunningGame.to(stateErrorGame)
    ClientRollDice = stateMyTurn.to(stateForkMyTurn)
    ServerPingPlayer = stateMyTurn.to(stateMyTurn)

    ResponseServerDiceEndTurn = stateForkMyTurn.to(stateRunningGame)
    ResponseServerDiceNext = stateForkMyTurn.to(stateNextDice)
    ClientEndTurn = stateNextDice.to(stateRunningGame)
    ClientNextDice = stateNextDice.to(stateMyTurn)
    ServerPingPlayer = stateNextDice.to(stateNextDice)
    ResponseServerNextDiceEndScore = stateForkNextDice.to(stateRunningGame)
    ResponseServerNextDiceSuccess = stateForkNextDice.to(stateNextDice)


state_machine = StateMachine()
G_game_state_machine = GameStateMachine(state_machine)

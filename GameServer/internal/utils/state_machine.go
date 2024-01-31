package utils

import (
	"gameserver/pkg/stateless"
)

/*var (
	instanceSM *stateless.stateMachine
	onceSM     sync.Once
)*/

var (
	stateEnd              = stateless.State("End")
	stateErrorGame        = stateless.State("ErrorGame")
	stateErrorLobby       = stateless.State("ErrorLobby")
	stateErrorRunningGame = stateless.State("ErrorRunning_Game")
	stateForkMyTurn       = stateless.State("ForkMyTurn")
	stateGame             = stateless.State("game")
	stateLobby            = stateless.State("Lobby")
	stateMyTurn           = stateless.State("MyTurn")
	stateNextDice         = stateless.State("NextDice")
	stateRunningGame      = stateless.State("Running_Game")
	stateStart            = stateless.State("Start")
	stateForkNextDice     = stateless.State("ForkNextDice")
)

/*func GetInstanceStateMachine() *stateless.stateMachine {
	onceSM.Do(func() {
		instanceSM = stateless.NewStateMachine(stateStart)
	})
	return instanceSM
}*/

func CreateStateMachine() *stateless.StateMachine {
	stateMachine := stateless.NewStateMachine(stateStart)
	initilize(stateMachine)
	return stateMachine
}

func initilize(stateMachine *stateless.StateMachine) {

	stateMachine.Configure(stateStart).
		Permit(CGCommands.ClientLogin.Trigger, stateLobby)

	stateMachine.Configure(stateLobby).
		Permit(CGCommands.ClientLogout.Trigger, stateEnd).
		Permit(CGCommands.ServerUpdateGameList.Trigger, stateLobby).
		Permit(CGCommands.ClientJoinGame.Trigger, stateGame).
		Permit(CGCommands.ClientCreateGame.Trigger, stateGame).
		Permit(CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorLobby)

	stateMachine.Configure(stateErrorLobby).
		Permit(CGCommands.ServerReconnectGameList.Trigger, stateLobby)

	stateMachine.Configure(stateGame).
		Permit(CGCommands.ClientStartGame.Trigger, stateRunningGame).
		Permit(CGCommands.ServerUpdateStartGame.Trigger, stateRunningGame).
		Permit(CGCommands.ServerUpdatePlayerList.Trigger, stateGame).
		Permit(CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorGame)

	stateMachine.Configure(stateErrorGame).
		Permit(CGCommands.ServerReconnectPlayerList.Trigger, stateGame).
		Permit(CGCommands.ServerUpdateStartGame.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateRunningGame).
		Permit(CGCommands.ServerUpdateEndScore.Trigger, stateLobby).
		Permit(CGCommands.ServerStartTurn.Trigger, stateMyTurn).
		Permit(CGCommands.ServerUpdateGameData.Trigger, stateRunningGame).
		Permit(CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateErrorRunningGame).
		Permit(CGCommands.ServerReconnectGameData.Trigger, stateRunningGame).
		Permit(CGCommands.ServerUpdateEndScore.Trigger, stateErrorGame)

	stateMachine.Configure(stateMyTurn).
		Permit(CGCommands.ClientRollDice.Trigger, stateForkMyTurn).
		Permit(CGCommands.ServerPingPlayer.Trigger, stateMyTurn).
		Permit(CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateForkMyTurn).
		Permit(CGCommands.ResponseServerDiceEndTurn.Trigger, stateRunningGame).
		Permit(CGCommands.ResponseServerDiceNext.Trigger, stateNextDice)

	stateMachine.Configure(stateNextDice).
		Permit(CGCommands.ClientEndTurn.Trigger, stateRunningGame).
		Permit(CGCommands.ClientNextDice.Trigger, stateMyTurn).
		Permit(CGCommands.ServerPingPlayer.Trigger, stateNextDice)

	stateMachine.Configure(stateForkNextDice).
		Permit(CGCommands.ResponseServerNextDiceEndScore.Trigger, stateRunningGame).
		Permit(CGCommands.ResponseServerNextDiceSuccess.Trigger, stateNextDice)
}

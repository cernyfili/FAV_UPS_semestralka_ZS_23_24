package models

import (
	"gameserver/internal/utils/constants"
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
		Permit(constants.CGCommands.ClientLogin.Trigger, stateLobby)

	stateMachine.Configure(stateLobby).
		Permit(constants.CGCommands.ClientLogout.Trigger, stateEnd).
		PermitReentry(constants.CGCommands.ServerUpdateGameList.Trigger).
		Permit(constants.CGCommands.ClientJoinGame.Trigger, stateGame).
		Permit(constants.CGCommands.ClientCreateGame.Trigger, stateGame).
		Permit(constants.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorLobby)

	stateMachine.Configure(stateErrorLobby).
		Permit(constants.CGCommands.ServerReconnectGameList.Trigger, stateLobby)

	stateMachine.Configure(stateGame).
		Permit(constants.CGCommands.ClientStartGame.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ServerUpdateStartGame.Trigger, stateRunningGame).
		PermitReentry(constants.CGCommands.ServerUpdatePlayerList.Trigger).
		Permit(constants.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorGame)

	stateMachine.Configure(stateErrorGame).
		Permit(constants.CGCommands.ServerReconnectPlayerList.Trigger, stateGame).
		Permit(constants.CGCommands.ServerUpdateStartGame.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateRunningGame).
		Permit(constants.CGCommands.ServerUpdateEndScore.Trigger, stateLobby).
		Permit(constants.CGCommands.ServerStartTurn.Trigger, stateMyTurn).
		PermitReentry(constants.CGCommands.ServerUpdateGameData.Trigger).
		Permit(constants.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateErrorRunningGame).
		Permit(constants.CGCommands.ServerReconnectGameData.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ServerUpdateEndScore.Trigger, stateErrorGame)

	stateMachine.Configure(stateMyTurn).
		Permit(constants.CGCommands.ClientRollDice.Trigger, stateForkMyTurn).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger).
		Permit(constants.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateForkMyTurn).
		Permit(constants.CGCommands.ResponseServerEndTurn.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ResponseServerSelectCubes.Trigger, stateNextDice)

	stateMachine.Configure(stateNextDice).
		Permit(constants.CGCommands.ClientEndTurn.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ClientSelectedCubes.Trigger, stateMyTurn).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger)

	stateMachine.Configure(stateForkNextDice).
		Permit(constants.CGCommands.ResponseServerEndScore.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ResponseServerDiceSuccess.Trigger, stateNextDice)
}

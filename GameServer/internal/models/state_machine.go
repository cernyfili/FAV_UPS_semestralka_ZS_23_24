package models

import (
	"gameserver/internal/utils"
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
		Permit(utils.CGCommands.ClientLogin.Trigger, stateLobby)

	stateMachine.Configure(stateLobby).
		Permit(utils.CGCommands.ClientLogout.Trigger, stateEnd).
		PermitReentry(utils.CGCommands.ServerUpdateGameList.Trigger).
		Permit(utils.CGCommands.ClientJoinGame.Trigger, stateGame).
		Permit(utils.CGCommands.ClientCreateGame.Trigger, stateGame).
		Permit(utils.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorLobby)

	stateMachine.Configure(stateErrorLobby).
		Permit(utils.CGCommands.ServerReconnectGameList.Trigger, stateLobby)

	stateMachine.Configure(stateGame).
		Permit(utils.CGCommands.ClientStartGame.Trigger, stateRunningGame).
		Permit(utils.CGCommands.ServerUpdateStartGame.Trigger, stateRunningGame).
		PermitReentry(utils.CGCommands.ServerUpdatePlayerList.Trigger).
		Permit(utils.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorGame)

	stateMachine.Configure(stateErrorGame).
		Permit(utils.CGCommands.ServerReconnectPlayerList.Trigger, stateGame).
		Permit(utils.CGCommands.ServerUpdateStartGame.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateRunningGame).
		Permit(utils.CGCommands.ServerUpdateEndScore.Trigger, stateLobby).
		Permit(utils.CGCommands.ServerStartTurn.Trigger, stateMyTurn).
		PermitReentry(utils.CGCommands.ServerUpdateGameData.Trigger).
		Permit(utils.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateErrorRunningGame).
		Permit(utils.CGCommands.ServerReconnectGameData.Trigger, stateRunningGame).
		Permit(utils.CGCommands.ServerUpdateEndScore.Trigger, stateErrorGame)

	stateMachine.Configure(stateMyTurn).
		Permit(utils.CGCommands.ClientRollDice.Trigger, stateForkMyTurn).
		PermitReentry(utils.CGCommands.ServerPingPlayer.Trigger).
		Permit(utils.CGCommands.ErrorPlayerUnreachable.Trigger, stateErrorRunningGame)

	stateMachine.Configure(stateForkMyTurn).
		Permit(utils.CGCommands.ResponseServerDiceEndTurn.Trigger, stateRunningGame).
		Permit(utils.CGCommands.ResponseServerDiceNext.Trigger, stateNextDice)

	stateMachine.Configure(stateNextDice).
		Permit(utils.CGCommands.ClientEndTurn.Trigger, stateRunningGame).
		Permit(utils.CGCommands.ClientNextDice.Trigger, stateMyTurn).
		PermitReentry(utils.CGCommands.ServerPingPlayer.Trigger)

	stateMachine.Configure(stateForkNextDice).
		Permit(utils.CGCommands.ResponseServerNextDiceEndScore.Trigger, stateRunningGame).
		Permit(utils.CGCommands.ResponseServerNextDiceSuccess.Trigger, stateNextDice)
}

package state_machine

import (
	"gameserver/internal/utils/constants"
	"gameserver/pkg/stateless"
)

/*var (
	instanceSM *stateless.stateMachine
	onceSM     sync.Once
)*/

type StateName struct {
	StateEnd          string
	StateForkMyTurn   string
	StateGame         string
	StateLobby        string
	StateMyTurn       string
	StateNextDice     string
	StateRunningGame  string
	StateStart        string
	StateForkNextDice string
	StateReconnect    string
}

var StateNameMap = StateName{
	StateEnd:          "End",
	StateForkMyTurn:   "ForkMyTurn",
	StateGame:         "Game",
	StateLobby:        "Lobby",
	StateMyTurn:       "MyTurn",
	StateNextDice:     "NextDice",
	StateRunningGame:  "Running_Game",
	StateStart:        "Start",
	StateForkNextDice: "ForkNextDice",
	StateReconnect:    "Reconnect",
}

var (
	stateEnd          = stateless.State(StateNameMap.StateEnd)
	stateForkMyTurn   = stateless.State(StateNameMap.StateForkMyTurn)
	stateGame         = stateless.State(StateNameMap.StateGame)
	stateLobby        = stateless.State(StateNameMap.StateLobby)
	stateMyTurn       = stateless.State(StateNameMap.StateMyTurn)
	stateNextDice     = stateless.State(StateNameMap.StateNextDice)
	stateRunningGame  = stateless.State(StateNameMap.StateRunningGame)
	stateStart        = stateless.State(StateNameMap.StateStart)
	stateForkNextDice = stateless.State(StateNameMap.StateForkNextDice)
	stateReconnect    = stateless.State(StateNameMap.StateReconnect)
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
		Permit(constants.CGCommands.ClientLogin.Trigger, stateLobby).
		Permit(constants.CGCommands.ClientReconnect.Trigger, stateReconnect)

	stateMachine.Configure(stateReconnect).
		Permit(constants.CGCommands.ResponseServerReconnectBeforeGame.Trigger, stateGame).
		Permit(constants.CGCommands.ResponseServerReconnectRunningGame.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ResponseServerGameList.Trigger, stateLobby).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger)

	stateMachine.Configure(stateLobby).
		Permit(constants.CGCommands.ClientLogout.Trigger, stateEnd).
		PermitReentry(constants.CGCommands.ServerUpdateGameList.Trigger).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger).
		Permit(constants.CGCommands.ClientJoinGame.Trigger, stateGame).
		Permit(constants.CGCommands.ClientCreateGame.Trigger, stateGame)

	stateMachine.Configure(stateGame).
		Permit(constants.CGCommands.ClientStartGame.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ServerUpdateStartGame.Trigger, stateRunningGame).
		PermitReentry(constants.CGCommands.ServerUpdatePlayerList.Trigger).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger)

	stateMachine.Configure(stateRunningGame).
		Permit(constants.CGCommands.ServerUpdateEndScore.Trigger, stateLobby).
		Permit(constants.CGCommands.ServerStartTurn.Trigger, stateMyTurn).
		Permit(constants.CGCommands.ServerUpdateNotEnoughPlayers.Trigger, stateLobby).
		PermitReentry(constants.CGCommands.ServerUpdateGameData.Trigger).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger)

	stateMachine.Configure(stateMyTurn).
		Permit(constants.CGCommands.ClientRollDice.Trigger, stateForkMyTurn).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger).
		PermitReentry(constants.CGCommands.ServerUpdateGameData.Trigger).
		Permit(constants.CGCommands.ServerUpdateNotEnoughPlayers.Trigger, stateLobby)

	stateMachine.Configure(stateForkMyTurn).
		Permit(constants.CGCommands.ResponseServerEndTurn.Trigger, stateRunningGame).
		Permit(constants.CGCommands.ResponseServerSelectCubes.Trigger, stateNextDice)

	stateMachine.Configure(stateNextDice).
		Permit(constants.CGCommands.ClientSelectedCubes.Trigger, stateForkNextDice).
		PermitReentry(constants.CGCommands.ServerPingPlayer.Trigger).
		PermitReentry(constants.CGCommands.ServerUpdateGameData.Trigger).
		Permit(constants.CGCommands.ServerUpdateNotEnoughPlayers.Trigger, stateLobby)

	stateMachine.Configure(stateForkNextDice).
		Permit(constants.CGCommands.ResponseServerEndScore.Trigger, stateLobby).
		Permit(constants.CGCommands.ResponseServerDiceSuccess.Trigger, stateMyTurn)
}

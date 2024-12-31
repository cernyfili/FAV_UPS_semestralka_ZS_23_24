package utils

import (
	"gameserver/pkg/stateless"
)

// region CONSTANTS
const (
	CMessageSignature    string = "KIVUPS"
	CMessageEndDelimiter        = '\n'
	CMaxMessageSize      int    = 1024
)

const CLogFilePath string = "logs/app.log"

var (
	CGScoreCubeValues = []ScoreCube{
		{1, 100},
		{5, 50},
	}
)

//endregion

// region DATA STRUCTURES
type Params struct {
	Name  string
	Value string
}

var CGNetworkEmptyParams []Params

type ScoreCube struct {
	Value      int
	ScoreValue int
}

type Command struct {
	CommandID   int
	Trigger     stateless.Trigger
	ParamsNames []string
}

type CommandType struct {
	ClientCreateGame Command
	ClientJoinGame   Command
	ClientLogin      Command
	ClientLogout     Command
	ClientRollDice   Command
	ClientStartGame  Command

	ErrorPlayerUnreachable Command

	ClientEndTurn         Command
	ClientNextDice        Command
	ResponseClientSuccess Command

	ResponseServerNextDiceEndScore    Command
	ResponseServerDiceEndTurn         Command
	ResponseServerDiceNext            Command
	ResponseServerSuccess             Command
	ResponseServerErrDuplicitNickname Command

	ServerPingPlayer          Command
	ServerUpdateEndScore      Command
	ServerUpdateStartGame     Command
	ServerReconnectGameData   Command
	ServerReconnectGameList   Command
	ServerReconnectPlayerList Command
	ServerStartTurn           Command
	ServerUpdateGameData      Command
	ServerUpdateGameList      Command
	ServerUpdatePlayerList    Command
	ResponseServerError       Command
	ResponseServerGameList    Command
	ClientReconnect           Command

	ResponseServerNextDiceSuccess Command
}

//endregion

var CGCommands = CommandType{

	//CLIENT->SERVER
	ClientLogin:      Command{1, stateless.Trigger("ClientLogin"), []string{""}},
	ClientCreateGame: Command{2, stateless.Trigger("ClientCreateGame"), []string{"gameName, maxPlayers"}},
	ClientJoinGame:   Command{3, stateless.Trigger("ClientJoinGame"), []string{"gameID"}},
	ClientStartGame:  Command{4, stateless.Trigger("ClientStartGame"), []string{""}},
	ClientRollDice:   Command{5, stateless.Trigger("ClientRollDice"), []string{""}},
	ClientLogout:     Command{7, stateless.Trigger("ClientLogout"), []string{""}},
	ClientReconnect:  Command{8, stateless.Trigger("ClientReconnect"), []string{""}},
	ClientNextDice:   Command{61, stateless.Trigger("ClientNextDice"), []string{""}},
	ClientEndTurn:    Command{62, stateless.Trigger("ClientEndTurn"), []string{""}},

	//RESPONSES SERVER->CLIENT
	ResponseServerSuccess:             Command{30, nil, []string{""}},
	ResponseServerErrDuplicitNickname: Command{31, nil, []string{""}},
	ResponseServerError:               Command{32, nil, []string{"message"}},
	ResponseServerGameList:            Command{33, nil, []string{"gameList"}},

	ResponseServerDiceNext:    Command{34, stateless.Trigger("ResponseServerDiceNext"), []string{""}},
	ResponseServerDiceEndTurn: Command{35, stateless.Trigger("ResponseServerDiceEndTurn"), []string{""}},

	ResponseServerNextDiceEndScore: Command{36, stateless.Trigger("ResponseServerNextDiceEndScore"), []string{""}},
	ResponseServerNextDiceSuccess:  Command{37, stateless.Trigger("ResponseServerNextDiceSuccess"), []string{""}},

	//SERVER->CLIENT
	ServerUpdateStartGame:  Command{41, stateless.Trigger("ServerUpdateStartGame"), []string{""}},
	ServerUpdateEndScore:   Command{42, stateless.Trigger("ServerUpdateEndScore"), []string{""}},
	ServerUpdateGameData:   Command{43, stateless.Trigger("ServerUpdateGameData"), []string{""}},
	ServerUpdateGameList:   Command{44, stateless.Trigger("ServerUpdateGameList"), []string{"gameList"}},
	ServerUpdatePlayerList: Command{45, stateless.Trigger("ServerUpdatePlayerList"), []string{""}},

	ServerReconnectGameList:   Command{46, stateless.Trigger("ServerReconnectGameList"), []string{""}},
	ServerReconnectGameData:   Command{47, stateless.Trigger("ServerReconnectGameData"), []string{""}},
	ServerReconnectPlayerList: Command{48, stateless.Trigger("ServerReconnectPlayerList"), []string{""}},

	ServerStartTurn:  Command{49, stateless.Trigger("ServerStartTurn"), []string{""}},
	ServerPingPlayer: Command{50, stateless.Trigger("ServerPingPlayer"), []string{""}},

	//RESPONSES CLIENT->SERVER
	ResponseClientSuccess: Command{60, stateless.Trigger("ResponseClientSuccess"), []string{""}},

	ErrorPlayerUnreachable: Command{70, stateless.Trigger("ErrorPlayerUnreachable"), []string{""}},
}

// region DATA STRUCTURES

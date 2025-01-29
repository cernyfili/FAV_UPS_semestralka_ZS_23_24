package constants

import (
	"gameserver/pkg/stateless"
	"reflect"
	"time"
)

// region CONSTANTS

type Brackets struct {
	Opening string
	Closing string
}

// region Message Constants
var (
	CParamsBrackets = Brackets{Opening: "{", Closing: "}"}
	CArrayBrackets  = Brackets{Opening: "[", Closing: "]"}
)

const (
	CMessageSignature string = "KIVUPS"

	CMessageEndDelimiter        = "\n"
	CParamsDelimiter            = ","
	CParamsListElementDelimiter = ";"
	CParamsKeyValueDelimiter    = ":"
	CParamsWrapper              = "\""

	CMessageMaxSize      int = 1024
	CMessageBufferSize   int = 1024
	CMessageNameMinChars int = 3
	CMessageNameMaxChars int = 20

	CMessageTimeFormat string = "2006-01-02 15:04:05.000000"
)

//endregion

// region Network Constants
var (
	CConIPadress string
	CConnPort    string
)

const (
	CConnType            = "tcp"
	CTimeout             = 5 * time.Second
	CPingTime            = 5 * time.Second
	CTotalDisconnectTime = 10 * time.Second
)

//endregion

//region FilePaths

const CLogsFolderPath string = "logs"
const CConfigFilePath string = "config.json"

//endregion

var (
	CGScoreCubeValues = []ScoreCube{
		{1, 100},
		{5, 50},
	}
)

var CGNetworkEmptyParams []Params

//endregion

// region DATA STRUCTURES
type Params struct {
	Name  string
	Value string
}

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
	//CLIENT->SERVER
	ClientCreateGame Command
	ClientJoinGame   Command
	ClientLogin      Command
	ClientLogout     Command
	ClientRollDice   Command
	ClientStartGame  Command

	ClientEndTurn       Command
	ClientSelectedCubes Command

	ClientReconnect Command

	//RESPONSES CLIENT->SERVER
	ResponseClientSuccess Command

	//RESPONSES SERVER->CLIENT
	ResponseServerEndScore    Command
	ResponseServerEndTurn     Command
	ResponseServerSelectCubes Command
	ResponseServerSuccess     Command

	ResponseServerError       Command
	ResponseServerGameList    Command
	ResponseServerDiceSuccess Command

	ResponseServerReconnectRunningGame Command
	ResponseServerReconnectBeforeGame  Command

	// SERVER->SINGLE CLIENT
	ServerPingPlayer Command
	ServerStartTurn  Command

	// SERVER->MULTIPLE CLIENTS - ONCE
	ServerUpdateEndScore         Command
	ServerUpdateStartGame        Command
	ServerUpdateNotEnoughPlayers Command

	// SERVER->MULTIPLE CLIENTS - CONTINUOUSLY
	ServerUpdateGameData   Command
	ServerUpdateGameList   Command
	ServerUpdatePlayerList Command
}

//endregion

var CGCommands = CommandType{

	//CLIENT->SERVER
	ClientLogin:      Command{1, stateless.Trigger("ClientLogin"), []string{""}},
	ClientCreateGame: Command{2, stateless.Trigger("ClientCreateGame"), []string{"gameName", "maxPlayers"}},
	ClientJoinGame:   Command{3, stateless.Trigger("ClientJoinGame"), []string{"gameName"}},
	ClientStartGame:  Command{4, stateless.Trigger("ClientStartGame"), []string{""}},
	ClientRollDice:   Command{5, stateless.Trigger("ClientRollDice"), []string{""}},
	ClientLogout:     Command{7, stateless.Trigger("ClientLogout"), []string{""}},

	ClientReconnect: Command{8, stateless.Trigger("ClientReconnect"), []string{""}},

	ClientSelectedCubes: Command{61, stateless.Trigger("ClientSelectedCubes"), []string{"cubeValues"}}, //check everywhere
	ClientEndTurn:       Command{62, stateless.Trigger("ClientEndTurn"), []string{""}},

	//RESPONSES SERVER->CLIENT
	ResponseServerSuccess: Command{30, nil, []string{""}},
	ResponseServerError:   Command{32, nil, []string{"message"}},

	ResponseServerGameList: Command{33, nil, []string{"gameList"}},

	ResponseServerSelectCubes: Command{34, stateless.Trigger("ResponseServerSelectCubes"), []string{"cubeValues"}},
	ResponseServerEndTurn:     Command{35, stateless.Trigger("ResponseServerEndTurn"), []string{""}},

	ResponseServerEndScore:    Command{36, stateless.Trigger("ResponseServerEndScore"), []string{""}},
	ResponseServerDiceSuccess: Command{37, stateless.Trigger("ResponseServerDiceSuccess"), []string{""}},

	ResponseServerReconnectBeforeGame:  Command{46, stateless.Trigger("ResponseServerReconnectBeforeGame"), []string{"gameList"}},
	ResponseServerReconnectRunningGame: Command{47, stateless.Trigger("ResponseServerReconnectRunningGame"), []string{"gameData"}},

	// SERVER->CLIENT
	//// SERVER -> ALL CLIENTS
	////// ONCE
	ServerUpdateStartGame:        Command{41, stateless.Trigger("ServerUpdateStartGame"), []string{""}},
	ServerUpdateEndScore:         Command{42, stateless.Trigger("ServerUpdateEndScore"), []string{"playerName"}},
	ServerUpdateNotEnoughPlayers: Command{51, stateless.Trigger("ServerUpdateNotEnoughPlayers"), []string{""}},

	////// CONTINUOUSLY
	ServerUpdateGameData:   Command{43, stateless.Trigger("ServerUpdateGameData"), []string{"gameData"}},
	ServerUpdateGameList:   Command{44, stateless.Trigger("ServerUpdateGameList"), []string{"gameList"}},
	ServerUpdatePlayerList: Command{45, stateless.Trigger("ServerUpdatePlayerList"), []string{"playerList"}},

	//// SERVER -> SINGLE CLIENT
	ServerStartTurn:  Command{49, stateless.Trigger("ServerStartTurn"), []string{""}},
	ServerPingPlayer: Command{50, stateless.Trigger("ServerPingPlayer"), []string{""}},

	//RESPONSES CLIENT->SERVER
	ResponseClientSuccess: Command{60, stateless.Trigger("ResponseClientSuccess"), []string{""}},
}

func GetCommandName(commandID int) string {
	v := reflect.ValueOf(CGCommands)
	for i := 0; i < v.NumField(); i++ {
		command := v.Field(i).Interface().(Command)
		if command.CommandID == commandID {
			//return the command name
			return v.Type().Field(i).Name
		}
	}
	return ""
}

func IsAlphaNumeric(name string) bool {
	for _, r := range name {
		//if is number
		if r >= '0' && r <= '9' {
			continue
		}
		//if is letter
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		return false
	}

	return true
}

// region GLOBAL VARIABLES

// region GLOBAL VARIABLES
const (
	CMaxScore = 100
)

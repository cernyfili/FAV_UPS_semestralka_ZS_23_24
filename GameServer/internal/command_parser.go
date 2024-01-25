package internal

import (
	"fmt"
	"gameserver/internal/utils"
	"net"
	"regexp"
)

//region GLOBAL VARIABLES

var gPlayerlist = utils.GetInstancePlayerList()

var gGamelist = utils.GetInstanceGameList()

var cCommands = CommandType{

	//CLIENT->SERVER
	ClientLogin:      Command{1},
	ClientCreateGame: Command{2},
	ClientJoinGame:   Command{3},
	ClientStartGame:  Command{4},
	ClientRollDice:   Command{5},
	ClientLogout:     Command{7},

	//RESPONSES SERVER->CLIENT
	ResponseServerSuccess:             Command{30},
	ResponseServerErrDuplicitNickname: Command{31},

	ResponseServerDiceNext:     Command{32},
	ResponseServerDiceEndTurn:  Command{33},
	ResponseServerDiceEndScore: Command{34},

	//SERVER->CLIENT
	ServerPlayerStartGame: Command{41},
	ServerPlayerEndScore:  Command{42},

	ServerUpdateGameData:   Command{43},
	ServerUpdateGameList:   Command{44},
	ServerUpdatePlayerList: Command{45},

	ServerReconnectGameList:   Command{46},
	ServerReconnectGameData:   Command{47},
	ServerReconnectPlayerList: Command{48},

	ServerStartTurn:  Command{49},
	ServerPingPlayer: Command{50},

	//RESPONSES CLIENT->SERVER
	ResponseClientSuccess: Command{60},

	ResponseClientNextDice: Command{61},
	ResponseClientEndTurn:  Command{62},
}

// CommandsHandlers is a map of valid commands and their information.
var CommandsHandlers = map[int]CommandInfo{
	//1: {"PLAYER_LOGIN", processPlayerLogin},
	cCommands.ClientCreateGame.CommandID: {processCreateGame},
	cCommands.ClientJoinGame.CommandID:   {processJoinGame},
	cCommands.ClientStartGame.CommandID:  {processStartGame},
	cCommands.ClientRollDice.CommandID:   {processRollDice}, //todo make start player turn and end player turn
	cCommands.ClientEndGame.CommandID:    {processEndGame},
	cCommands.ClientLogout.CommandID:     {processPlayerLogout},
}

//endregion

//region STRUCTURES

// CommandInfo represents information about a command.
type CommandInfo struct {
	Handler func(player utils.Player, args string) error //todo should all return network message format
}

type Command struct {
	CommandID int
}

type CommandType struct {
	ClientCreateGame Command
	ClientJoinGame   Command
	ClientLogin      Command
	ClientLogout     Command
	ClientRollDice   Command
	ClientStartGame  Command

	ErrorPlayerUnreachable Command

	ResponseClientEndTurn  Command
	ResponseClientNextDice Command

	ResponseServerDiceEndScore Command
	ResponseServerDiceEndTurn  Command
	ResponseServerDiceNext     Command

	ServerPingPlayer Command

	ServerPlayerEndScore  Command
	ServerPlayerStartGame Command

	ServerReconnectGameData   Command
	ServerReconnectGameList   Command
	ServerReconnectPlayerList Command

	ServerStartTurn Command

	ServerUpdateGameData              Command
	ServerUpdateGameList              Command
	ServerUpdatePlayerList            Command
	ResponseServerSuccess             Command
	ResponseServerErrDuplicitNickname Command
	ResponseClientSuccess             Command
}

//endregion

func ProcessMessage(message utils.Message, conn net.Conn) error {

	//Check if valid signature
	if message.Signature != utils.CMessageSignature {
		return fmt.Errorf("invalid signature")
	}

	commandID := message.CommandID
	playerNickname := message.PlayerNickname
	timeStamp := message.TimeStamp
	params := message.Parameters

	//check if values valid
	if !isValidCommand(commandID) {
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}
	if !isValidNickname(playerNickname) {
		return fmt.Errorf("invalid nickname")
	}

	connectionInfo := utils.ConnectionInfo{
		Connection: conn,
		TimeStamp:  timeStamp,
	}

	//If player_login
	if commandID == 1 {
		err := processPlayerLogin(playerNickname, connectionInfo)
		if err != nil {
			return err
		}
	}

	//Get player
	player, err := gPlayerlist.GetItem(playerNickname)
	if err != nil {
		return err
	}

	//Set connection to player
	player.ConnectionInfo = connectionInfo

	// Call the corresponding handler function
	info := CommandsHandlers[commandID]
	err = info.Handler(*player, params)
	if err != nil {
		return err
	}

	return nil
}

func isValidNickname(nickname string) bool {
	//Check base on regex [A-Za-z0-9_\\-]+

	// Define the regular expression pattern to match the args
	pattern := "[A-Za-z0-9_\\-]+"

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the matches in the string
	matches := regex.FindStringSubmatch(nickname)

	// Check if there is a match and extract the nickname
	if len(matches) != 1 || nickname == "" {
		return false
	}

	return true
}

func isValidCommand(command int) bool {
	_, exists := CommandsHandlers[command]
	if !exists {
		// The command is not valid.
		return false
	}

	// Check if the number of arguments matches the expected count.
	return true
}

//endregion

//region Processing functions

func processPlayerLogin(playerNickname string, connectionInfo utils.ConnectionInfo) error {
	responseInfo := utils.NetworkResponseInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	//Check if playerNickname in list
	_, err := gPlayerlist.GetItem(playerNickname)
	if err != nil {
		err = SendResponseDuplicitNickname(responseInfo)
		return fmt.Errorf("playerNickname already in game %s", err)
	}

	// Add the player to the playerData
	player := utils.Player{
		Nickname:       playerNickname,
		IsConnected:    true,
		Game:           nil,
		ConnectionInfo: connectionInfo,
	}

	err = gPlayerlist.AddItem(playerNickname, &player)
	if err != nil {
		return fmt.Errorf("Error adding player: %s", err)
	}

	// Send the response
	err = SendResponseSuccess(responseInfo)
	if err != nil {
		return fmt.Errorf("Error sending response: %s", err)
	}

	return nil
}

/*
func processCreateGame(player utils.Player, args string) error {
	//todo check if already player in some game -> throw error
	//todo send error to player
	//todo get player id
	//todo save ip address to player
	//todo create thread for game

	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		return
	}

	// Parse the arguments
	argsStr := args[0]
	name, maxPlayers, err := ParseCreateGameArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		return
	}

	// Create the game
	game, err := utils.ClientCreateGame(name, maxPlayers)
	if err != nil {
		fmt.Println("Error creating game:", err)
		return
	}

	//add player to game
	err := game.AddPlayer(player)
	if err != nil {
		fmt.Println("Error adding player to game:", err)
		return
	}

	gameID, err := gGamelist.AddItem(game)
	if err != nil {
		fmt.Println("Error adding game:", err)
		return
	}

	// Send the response
	err = SendCreateGameResponse(gameID, *game)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processJoinGame(player utils.Player, args string) error {
	//todo check if already player in some game -> throw error
	//todo send error to player
	//todo send error if game is full
	//todo send error if game already started
	//todo send error if game doesn't exist
	//todo send error if game ended
	//todo MULTITHRED - connect player to thread -> channel data

	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		return
	}

	// Parse the arguments
	argsStr := args[0]
	gameID, playerID, err := ParseJoinGameArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		return
	}

	// Add the player to the game
	game, err := gGamelist.GetItem(gameID)
	if err != nil {
		fmt.Println("Error adding player:", err)
		return
	}

	playerItem, err := gPlayerlist.GetItem(playerID)
	if err != nil {
		fmt.Println("Error adding player:", err)
		return
	}

	player := playerItem.(*utils.Player)

	err = game.AddPlayer(player)
	if err != nil {
		fmt.Println("Error adding player:", err)
		return
	}

	// Send the response
	err = SendJoinGameResponse(gameID, *game)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processStartGame(player utils.Player, args string) error {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, fmt.Errorf("Invalid number of arguments")) //todo get response IP
		return
	}

	// Parse the arguments
	argsStr := args[0]
	gameID, playerID, err := ParseStartGameArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		SendErrorToPlayer(0, err) //todo get response IP
		return
	}

	//ERROR: player doesn't exist
	playerItem, err := gPlayerlist.GetItem(playerID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}
	player := playerItem.(*utils.Player)

	//ERROR: game doesn't exist
	game, err := gGamelist.GetItem(gameID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game already started
	if game.GameStateValue != utils.Created {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: player not in game
	if !game.HasPlayer(player) && player.Game != game {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	// Change the game state
	game.GameStateValue = utils.Running

	// Send the response
	err = SendStartGameResponse(gameID, *game) //todo send to all players in game
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processRollDice(player utils.Player, args string) error {
	//todo make it game based
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, fmt.Errorf("Invalid number of arguments")) //todo get response IP
		return
	}

	// Parse the arguments
	argsStr := args[0]
	playerID, err := ParseRollDiceArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		SendErrorToPlayer(0, err) //todo get response IP
		return
	}

	//ERROR: player doesn't exist
	playerItem, err := gPlayerlist.GetItem(playerID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}
	player := playerItem.(*utils.Player)

	//ERROR: player doesn't have a game
	game := player.Game
	hasGame := gGamelist.HasValue(game)
	if !hasGame {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game is not running
	if game.GameStateValue != utils.Running {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	// Roll the dice
	//todo make it based on rules (multiple rounds of one turn for one player)
	rollDiceResult = rollDice()

	// Send the response
	err = SendRollDiceResponse(gameID, *game) //todo send to all players in game
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func rollDice() int {
	//todo game logic
	//returns random number from
	return 0
}

func processPlayerLogout(player utils.Player, args string) error {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, fmt.Errorf("Invalid number of arguments")) //todo get response IP
		return
	}

	// Parse the arguments
	argsStr := args[0]
	playerID, err := ParsePlayerLogoutArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		SendErrorToPlayer(0, err) //todo get response IP
		return
	}

	//ERROR: player doesn't exist
	playerItem, err := gPlayerlist.GetItem(playerID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}
	player := playerItem.(*utils.Player)
	game := player.Game

	//ERROR: game doesn't exist
	hasGame := gGamelist.HasValue(game)
	if !hasGame {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	if !game.HasPlayer(player) {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game is not running
	if game.GameStateValue != utils.Running {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	game.RemovePlayer(player)

	// Send the response
	err = SendPlayerLogoutResponse(gameID, *game)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processEndGame(player utils.Player, args string) error {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, fmt.Errorf("Invalid number of arguments")) //todo get response IP
		return
	}

	// Parse the arguments
	argsStr := args[0]
	_, playerID, err := ParseEndGameArgs(argsStr) //todo only for playerID
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		SendErrorToPlayer(0, err) //todo get response IP
		return
	}

	//ERROR: player doesn't exist
	playerItem, err := gPlayerlist.GetItem(playerID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}
	player := playerItem.(*utils.Player)
	game := player.Game

	//ERROR: game doesn't exist
	hasGame := gGamelist.HasValue(game)
	if !hasGame {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	if !game.HasPlayer(player) {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game is not running
	if game.GameStateValue != utils.Running {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	game.GameStateValue = utils.Ended

	// Send the response
	err = SendEndGameResponse(gameID, *game) //todo send to all players in game
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

//endregion
*/

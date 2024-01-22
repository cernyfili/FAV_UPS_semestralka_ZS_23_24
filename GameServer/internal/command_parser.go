package internal

import (
	"errors"
	"fmt"
	"internal/data_structure"
	"strconv"
)

//region GLOBAL VARIABLES

var gPlayerlist = data_structure.CreateList() //todo multithread safe mutex - singleton

var gGamelist = data_structure.CreateGameList() //todo multithread safe mutex - singleton

//endregion

// CommandInfo represents information about a command.
type CommandInfo struct {
	CommandText string
	Handler     func([]string)
}

// Commands is a map of valid commands and their information.
var Commands = map[int]CommandInfo{
	1: {"PLAYER_LOGIN", processPlayerLogin},
	2: {"CREATE_GAME", processCreateGame},
	3: {"JOIN_GAME", processJoinGame},
	4: {"START_GAME", processStartGame},
	5: {"ROLL_DICE", processRollDice}, //todo make start player turn and end player turn
	6: {"END_GAME", processEndGame},
	7: {"PLAYER_LOGOUT", processPlayerLogout},
}

func main() {
	command, arguments := ReadCommand()

	commandNum, err := strconv.Atoi(command)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		return
	}
	//list where first element is string arguments
	// Create a list where the first element is a string argument
	argsArr := []string{arguments}

	if isValidCommand(commandNum) {
		fmt.Println("Valid command with appropriate arguments:", command, arguments)
		ProcessCommand(commandNum, argsArr) //todo add as first argument player pointer
	} else {
		fmt.Println("Invalid command or incorrect number of arguments:", command, arguments)
		// You can also return an error here if needed.
		// For example: return errors.New("Invalid command or incorrect number of arguments")
	}
}

// region Command functions

func ProcessCommand(command int, args []string) {
	info, exists := Commands[command]
	if !exists {
		fmt.Println("Unsupported command:", command)
		return
	}

	// Call the corresponding handler function
	info.Handler(args)
}

func isValidCommand(command int) bool {
	_, exists := Commands[command]
	if !exists {
		// The command is not valid.
		return false
	}

	// Check if the number of arguments matches the expected count.
	return true
}

func ReadCommand() (string, string) {
	command := "0"
	args := "test"
	// Now you can process the command and its arguments.
	return command, args
}

//endregion

//region Processing functions

func processPlayerLogin(args []string) {
	//todo check if already player in some game -> connect to that game
	//todo if error send to player error
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		return
	}

	// Parse the arguments
	argsStr := args[0]
	nickname, err := ParsePlayerLoginArgs(argsStr)
	if err != nil {
		fmt.Println("Error parsing command arguments:", err)
		return
	}

	// Add the player to the playerData
	player := data_structure.Player{
		Nickname:    nickname,
		IsConnected: true,
		Game:        nil,
	}

	playerID, err := gPlayerlist.AddItem(player)
	if err != nil {
		fmt.Println("Error adding player:", err)
		return
	}

	// Send the response
	err = SendPlayerConnectResponse(playerID, *gGamelist)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processCreateGame(args []string) {
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
	game, err := data_structure.CreateGame(name, maxPlayers)
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

func processJoinGame(args []string) {
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

	player := playerItem.(*data_structure.Player)

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

func processStartGame(args []string) {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, errors.New("Invalid number of arguments")) //todo get response IP
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
	player := playerItem.(*data_structure.Player)

	//ERROR: game doesn't exist
	game, err := gGamelist.GetItem(gameID)
	if err != nil {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game already started
	if game.GameStateValue != data_structure.Created {
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
	game.GameStateValue = data_structure.Running

	// Send the response
	err = SendStartGameResponse(gameID, *game) //todo send to all players in game
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func processRollDice(args []string) {
	//todo make it game based
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, errors.New("Invalid number of arguments")) //todo get response IP
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
	player := playerItem.(*data_structure.Player)

	//ERROR: player doesn't have a game
	game := player.Game
	hasGame := gGamelist.HasValue(game)
	if !hasGame {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	//ERROR: game is not running
	if game.GameStateValue != data_structure.Running {
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

func processPlayerLogout(args []string) {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, errors.New("Invalid number of arguments")) //todo get response IP
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
	player := playerItem.(*data_structure.Player)
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
	if game.GameStateValue != data_structure.Running {
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

func processEndGame(args []string) {
	//check arguments
	if len(args) != 1 {
		fmt.Println("Invalid number of arguments:", len(args))
		SendErrorToPlayer(0, errors.New("Invalid number of arguments")) //todo get response IP
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
	player := playerItem.(*data_structure.Player)
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
	if game.GameStateValue != data_structure.Running {
		fmt.Println("Error starting game:", err)
		SendErrorToPlayer(playerID, err)
		return
	}

	game.GameStateValue = data_structure.Ended

	// Send the response
	err = SendEndGameResponse(gameID, *game) //todo send to all players in game
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

//endregion

package internal

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/utils"
	"gameserver/internal/utils/errorHandeling"
	"net"
	"regexp"
	"strconv"
)

//region GLOBAL VARIABLES

var gPlayerlist = models.GetInstancePlayerList()

var gGamelist = models.GetInstanceGameList()

const (
	cMaxScore = 10000
)

// CommandsHandlers is a map of valid commands and their information.
var CommandsHandlers = map[int]CommandInfo{
	//1: {"PLAYER_LOGIN", processPlayerLogin},
	utils.CGCommands.ClientCreateGame.CommandID: {processClientCreateGame, utils.CGCommands.ClientCreateGame},
	utils.CGCommands.ClientJoinGame.CommandID:   {processClientJoinGame, utils.CGCommands.ClientJoinGame},
	utils.CGCommands.ClientStartGame.CommandID:  {processClientStartGame, utils.CGCommands.ClientStartGame},
	utils.CGCommands.ClientLogout.CommandID:     {processClientPlayerLogout, utils.CGCommands.ClientLogout},
	utils.CGCommands.ClientReconnect.CommandID:  {processClientReconnect, utils.CGCommands.ClientReconnect},
	//todo probably missing some commands which can come from client
}

//endregion

//region STRUCTURES

// CommandInfo represents information about a command.
type CommandInfo struct {
	Handler func(player *models.Player, params []utils.Params, command utils.Command) error //todo should all return network message format
	Command utils.Command
}

//endregion

func ProcessMessage(message models.Message, conn net.Conn) error {

	logger.Log.Debugf("Received message: %v", message)

	//Check if valid signature
	if message.Signature != utils.CMessageSignature {
		return fmt.Errorf("invalid signature")
	}

	commandID := message.CommandID
	playerNickname := message.PlayerNickname
	timeStamp := message.TimeStamp
	params := message.Parameters

	connectionInfo := models.ConnectionInfo{
		Connection: conn,
		TimeStamp:  timeStamp,
	}

	//SPECIAL CASE: If player_login
	if commandID == utils.CGCommands.ClientLogin.CommandID {
		if !isParamsEmpty(params) {
			errorHandeling.PrintError(fmt.Errorf("invalid number of arguments"))
			return fmt.Errorf("invalid number of arguments")
		}
		err := processPlayerLogin(playerNickname, connectionInfo, utils.CGCommands.ClientLogin)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Get player
	player, err := gPlayerlist.GetItem(playerNickname)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}

	//Set connection to player
	player.SetConnectionInfo(connectionInfo)

	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	//SPECIAL CASE: check if commandID valid
	commandInfo, err := getCommandInfo(commandID)
	if err != nil {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// SPECIAL CASE: check params are valid
	if !isValidParams(params, commandInfo.Command.ParamsNames) {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Call the corresponding handler function
	err = commandInfo.Handler(player, params, commandInfo.Command)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}

	return nil
}

func isParamsEmpty(params []utils.Params) bool {
	if len(params) == 0 {
		return true
	}
	return false
}

func isValidParams(params []utils.Params, names []string) bool {
	if len(params) != len(names) {
		return false
	}

	for i := 0; i < len(params); i++ {
		if params[i].Name != names[i] {
			return false
		}
	}

	return true
}

// region UTILS FUNCTIONS
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

func getCommandInfo(commandID int) (CommandInfo, error) {
	commandInfo, exists := CommandsHandlers[commandID]
	if !exists {
		// The commandID is not valid.
		return CommandInfo{}, fmt.Errorf("invalid commandID")
	}

	// Check if the number of arguments matches the expected count.
	return commandInfo, nil
}

func removePlayerFromLists(player *models.Player) error {

	//Remove player from playerlist
	err := gPlayerlist.RemoveItem(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}

	//Remove player from game
	game := player.GetGame()
	if game == nil {
		return nil
	}

	gameFromList, err := gGamelist.GetItem(game.GetGameID())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}

	if !gameFromList.HasPlayer(player) {
		return fmt.Errorf("cannot create game %w", err)
	}

	err = game.RemovePlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}

	return nil
}

func ConvertParamClientCreateGame(params []utils.Params, names []string) (string, int, error) {
	gameName := ""
	maxPlayers := 0

	if len(params) != len(names) {
		return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "gameName":
			gameName = current.Value
		case "maxPlayers":
			maxPlayers, err := strconv.Atoi(current.Value)
			if err != nil {
				errorHandeling.PrintError(err)
				return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
			}
		default:
			return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
		}
	}

	return gameName, maxPlayers, nil
}

func ConvertParamClientJoinGame(params []utils.Params, names []string) (int, error) {
	gameID := 0

	if len(params) != len(names) {
		return gameID, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return gameID, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "gameID":
			gameID, err := strconv.Atoi(current.Value)
			if err != nil {
				errorHandeling.PrintError(err)
				return gameID, fmt.Errorf("invalid number of arguments")
			}
		default:
			return gameID, fmt.Errorf("invalid number of arguments")
		}
	}

	return gameID, nil
}

func ConvertParamClientNextDice(params []utils.Params, names []string) ([]int, error) {
	var cubeValueIndex []int

	if len(params) != len(names) {
		return cubeValueIndex, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return cubeValueIndex, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "cubeValueIndex":
			array, err := ParseParamValueArray(current.Value)
			if err != nil {
				errorHandeling.PrintError(err)
				return cubeValueIndex, fmt.Errorf("invalid number of arguments")
			}
			for _, value := range array {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					errorHandeling.PrintError(err)
					return cubeValueIndex, fmt.Errorf("invalid number of arguments")
				}
				cubeValueIndex = append(cubeValueIndex, intValue)
			}
		default:
			return cubeValueIndex, fmt.Errorf("invalid number of arguments")
		}
	}

	return cubeValueIndex, nil
}

func addPlayerToGame(player *models.Player, err error, game *models.Game) error {
	// Add the player to the game
	err = game.AddPlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding player to game: %w", err)
	}

	// Set player game
	player.SetGame(game)
	return nil
}

func dissconectPlayer(player *models.Player, responseInfo models.NetworkResponseInfo) error {
	err := RemovePlayerFromGame(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	//send response
	err = SendResponseServerError(responseInfo, fmt.Errorf("You have been disconnected"))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	return nil
}

func RemovePlayerFromGame(player *models.Player) error {
	err := removePlayerFromLists(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Set player game
	player.SetGame(nil)
	return nil
}

//endregion

//endregion

//region Processing functions

// region Process functions
func processPlayerLogin(playerNickname string, connectionInfo models.ConnectionInfo, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	//Check if playerNickname in list
	if gPlayerlist.HasItem(playerNickname) {
		err := SendResponseServerErrDuplicitNickname(responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Send response
	isSuccess, err := SendResponseServerGameList(responseInfo, gGamelist.GetValuesArray())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	if !isSuccess {
		return nil
	}

	// Add the player to the playerData
	player := models.CreatePlayer(playerNickname, nil, true, connectionInfo)

	err = gPlayerlist.AddItem(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding player: %w", err)
	}

	// Move state machine to lobby
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func processClientCreateGame(player *models.Player, params []utils.Params, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}
	if !canFire {
		return dissconectPlayer(player, responseInfo)
	}

	//Convert params
	name, maxPlayers, err := ConvertParamClientCreateGame(params, command.ParamsNames)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error parsing command arguments: %w", err)
	}

	// Create the game
	game, err := models.CreateGame(name, maxPlayers)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error creating game: %w", err)
	}

	//add player to game
	err = game.AddPlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding player to game: %w", err)
	}

	//Set player game
	player.SetGame(game)

	_, err = gGamelist.AddItem(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding game: %w", err)
	}

	// Send the response
	err = SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Send update to all players
	err = CommunicationServerUpdateGameList(gPlayerlist.GetValuesArray(), gGamelist.GetValuesArray())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Change state machine
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func processClientJoinGame(player *models.Player, params []utils.Params, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Convert params
	gameID, err := ConvertParamClientJoinGame(params, command.ParamsNames) // assuming gameID is the first parameter
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error parsing command arguments: %w", err)
	}

	// Get the game
	game, err := gGamelist.GetItem(gameID)
	if err != nil {
		errorHandeling.PrintError(err)
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Check if player is already in game
	if player.GetGame() != nil {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Check if the game is full
	if game.IsFull() {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Check if the game has already started
	if game.GetState() != models.Created {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Add the player to the game
	err = addPlayerToGame(player, err, game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding player to game: %w", err)
	}

	// Send the response
	err = SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Send update to all players
	err = CommunicationServerUpdatePlayerList(game.GetPlayers())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Change state machine
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func processClientStartGame(player *models.Player, params []utils.Params, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//State machine check
	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//region LOGIC

	// Check if player got game
	game := player.GetGame()
	if game == nil {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Check if the game has already started
	if game.GetState() != models.Created {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Check if minimum players
	if game.IsEnoughPlayers() {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	game.SetState(models.Running)

	// Send the response
	err = SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Send update to all players
	err = CommunicationServerUpdateStartGame(game.GetPlayers())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Change state machine to Running_game
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//region Running_Game

	//send ServerUpdateGameData
	err = sendCurrentServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//region TURN logic
	//Running_Game --ServerStartTurn--> my_turn
	var isNextPlayerTurn = true
	for isNextPlayerTurn {
		isNextPlayerTurn, err = startPlayerTurn(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
	}

	//endregion

	//endregion

	//endregion

	return nil
}

func processClientPlayerLogout(player *models.Player, params []utils.Params, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Convert params
	if len(params) != 0 {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//region LOGIC
	err = RemovePlayerFromGame(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	// Send the response
	err = SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Change state machine
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func processClientReconnect(player *models.Player, params []utils.Params, command utils.Command) error {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Convert params
	if len(params) != 0 {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//region LOGIC

	wasHandled := false
	//ReconnectGameList
	canFire, err = player.GetStateMachine().CanFire(utils.CGCommands.ServerReconnectGameList.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if canFire {
		_, err := CommunicationServerReconnectGameList(player, gGamelist.GetValuesArray())

		err = player.FireStateMachine(utils.CGCommands.ServerReconnectGameList.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	//ReconnectGameData
	canFire, err = player.GetStateMachine().CanFire(utils.CGCommands.ServerReconnectGameData.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if canFire && !wasHandled {
		gameData, err := player.GetGame().GetGameData()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		_, err = CommunicationServerReconnectGameData(player, gameData)

		err = player.FireStateMachine(utils.CGCommands.ServerReconnectGameData.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	//ReconnectPlayerList
	canFire, err = player.GetStateMachine().CanFire(utils.CGCommands.ServerReconnectPlayerList.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}

	if canFire && !wasHandled {
		_, err := CommunicationServerReconnectPlayerList(player, player.GetGame().GetPlayers())

		err = player.FireStateMachine(utils.CGCommands.ServerReconnectPlayerList.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	if !wasHandled {
		err = dissconectPlayer(player, responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//endregion

	return nil
}

//endregion

func sendCurrentServerUpdateGameData(game *models.Game) error {
	gameData, err := game.GetGameData()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	err = CommunicationServerUpdateGameData(gameData, game.GetPlayers())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	return nil
}

func startPlayerTurn(game *models.Game) (bool, error) {
	gameData, err := game.GetGameData()
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	turnPlayer := gameData.TurnPlayer
	isSuccess, err := CommunicationServerStartTurn(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	if !isSuccess {
		//next Player turn
		return true, nil
	}

	err = turnPlayer.FireStateMachine(utils.CGCommands.ServerStartTurn.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	//My turn --ClientRollDice--> fork_my_turn
	message, isSuccess, err := CommunicationReadClientRollDice(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	if !isSuccess {
		//next Player turn
		return true, nil
	}

	err = turnPlayer.FireStateMachine(utils.CGCommands.ClientRollDice.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	//fork_my_turn --> ResponseServerDiceEndTurn
	//             --> ResponseServerDiceNext
	cubeValues, err := game.NewThrow(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	canBePlayed := CubesCanBePlayed(cubeValues)
	if !canBePlayed {
		//--> ResponseServerDiceEndTurn
		err = SendResponseServerDiceEndTurn(cubeValues, turnPlayer, message)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		//ServerUpdateGameData
		err = sendCurrentServerUpdateGameData(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		err = turnPlayer.FireStateMachine(utils.CGCommands.ResponseServerDiceEndTurn.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		// next Player turn
		return true, nil
	}

	//-->ResponseServerDiceNext
	err = SendResponseServerDiceNext(cubeValues, turnPlayer, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	err = turnPlayer.FireStateMachine(utils.CGCommands.ResponseServerDiceNext.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	//Next_dice --> Running_Game
	//           --> Fork_next_dice
	message, isSuccess, err = CommunicationReadClient(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	if !isSuccess {
		//next Player turn
		return true, nil
	}

	if message.CommandID == utils.CGCommands.ClientEndTurn.CommandID {
		//--> Running_Game

		//Send Success
		responseInfo := models.NetworkResponseInfo{
			ConnectionInfo: models.ConnectionInfo{
				Connection: turnPlayer.GetConnectionInfo().Connection,
				TimeStamp:  message.TimeStamp,
			},
			PlayerNickname: message.PlayerNickname,
		}
		err = SendResponseServerSuccess(responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		//Send GameUpdate
		err = sendCurrentServerUpdateGameData(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		err = turnPlayer.FireStateMachine(utils.CGCommands.ClientEndTurn.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		//next Player turn
		return true, nil
	}

	//--> Fork_next_dice
	if message.CommandID != utils.CGCommands.ClientNextDice.CommandID {
		return false, fmt.Errorf("Error sending response")
	}

	err = turnPlayer.FireStateMachine(utils.CGCommands.ClientNextDice.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	//Fork_next_dice --> ResponseServerNextDiceEndScore
	//               --> ResponseServerNextDiceSuccess
	selectedCubesIndexes, err := ConvertParamClientNextDice(message.Parameters, utils.CGCommands.ClientNextDice.ParamsNames)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	scoreIncrease, err := game.GetScoreIncrease(selectedCubesIndexes, turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	currentScore, err := game.GetPlayerScore(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	score := currentScore + scoreIncrease

	if score >= cMaxScore {
		//--> ResponseServerNextDiceEndScore

		err = SendResponseServerNextDiceEndScore(turnPlayer)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		//change player score
		err = game.SetPlayerScore(turnPlayer, score)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		//Send GameUpdate
		err = sendCurrentServerUpdateGameData(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		err = turnPlayer.FireStateMachine(utils.CGCommands.ResponseServerNextDiceEndScore.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		return false, nil
	}

	//--> ResponseServerNextDiceSuccess
	err = SendResponseServerNextDiceSuccess(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	err = turnPlayer.FireStateMachine(utils.CGCommands.ResponseServerNextDiceSuccess.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	// next Player turn
	return true, nil
}

func CubesCanBePlayed(values []int) bool {
	for _, value := range values {
		for _, score := range utils.CGScoreCubeValues {
			if value == score.Value {
				return true
			}
		}
	}
	return false
}

//endregion

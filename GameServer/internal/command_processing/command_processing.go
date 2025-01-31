package command_processing

import (
	"fmt"
	"gameserver/internal/command_processing/command_processing_utils"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/models/state_machine"
	"gameserver/internal/network"
	"gameserver/internal/parser"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/internal/utils/helpers"
	"net"
	"time"
)

// CommandsHandlers is a map of valid commands and their information.
var CommandsHandlers = map[int]CommandInfo{
	//1: {"PLAYER_LOGIN", processPlayerLogin} first case in if in ProcessMessage
	constants.CGCommands.ClientCreateGame.CommandID: {processClientCreateGame, constants.CGCommands.ClientCreateGame},
	constants.CGCommands.ClientJoinGame.CommandID:   {processClientJoinGame, constants.CGCommands.ClientJoinGame},
	constants.CGCommands.ClientStartGame.CommandID:  {processClientStartGame, constants.CGCommands.ClientStartGame},
	constants.CGCommands.ClientLogout.CommandID:     {processClientPlayerLogout, constants.CGCommands.ClientLogout},
	constants.CGCommands.ClientReconnect.CommandID:  {processClientReconnect, constants.CGCommands.ClientReconnect},
	//constants.CGCommands.ResponseClientSuccess.CommandID: {processResponseClientSucess, constants.CGCommands.ResponseClientSuccess},
	constants.CGCommands.ClientRollDice.CommandID:      {processClientRollDice, constants.CGCommands.ClientRollDice},
	constants.CGCommands.ClientSelectedCubes.CommandID: {processClientSelectedCubes, constants.CGCommands.ClientSelectedCubes},
}

//endregion

//region STRUCTURES

// CommandInfo represents information about a command.
type CommandInfo struct {
	Handler func(player *models.Player, params []constants.Params, command constants.Command) error
	Command constants.Command
}

//endregion

func ProcessMessage(message models.Message, conn net.Conn) error {
	logger.Log.Debugf("Starting to process: %v", message)

	// nested function handle invalid message format
	handleInvalidMessageFormat := func(player *models.Player, responseInfo models.MessageInfo) error {
		logger.Log.Errorf("Invalid message format")
		err := dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending response: %w", err)
		}
		errorHandeling.PrintError(fmt.Errorf("invalid command or incorrect number of arguments"))
		return nil
	}

	logger.Log.Debugf("Received message: %v", message)

	//Check if valid signature
	if message.Signature != constants.CMessageSignature {
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
	if commandID == constants.CGCommands.ClientLogin.CommandID {
		if !isParamsEmpty(params) {
			errorHandeling.PrintError(fmt.Errorf("invalid number of arguments"))
			return fmt.Errorf("invalid number of arguments")
		}
		err := processPlayerLogin(playerNickname, connectionInfo, constants.CGCommands.ClientLogin)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Get player
	player, err := models.GetInstancePlayerList().GetItem(playerNickname)
	if err != nil {
		logger.Log.Errorf("Error getting player: %v", err)
		errorHandeling.PrintError(err)
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}
	if player == nil {
		logger.Log.Errorf("Error player is nil")
		//close connection
		err := network.CloseConnection(conn)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error closing connection: %w", err)
		}
	}

	//Set connection to player
	player.SetConnectionInfo(connectionInfo)

	responseInfo := models.MessageInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	// SPECIAL CASE: Response Success
	if commandID == constants.CGCommands.ResponseClientSuccess.CommandID {
		err = processResponseClientSucess(player, timeStamp)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("invalid command or incorrect number of arguments")
		}
		return nil
	}

	commandInfo, err := getCommandInfo(commandID)

	//SPECIAL CASE: check if commandID valid
	if err != nil {
		return handleInvalidMessageFormat(player, responseInfo)
	}

	// SPECIAL CASE: check params are valid
	if !isValidParamsNames(params, commandInfo.Command.ParamsNames) {
		return handleInvalidMessageFormat(player, responseInfo)
	}

	// Call the corresponding handler function
	err = commandInfo.Handler(player, params, commandInfo.Command)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}

	return nil
}

// region UTILS FUNCTIONS

func isParamsEmpty(params []constants.Params) bool {
	if len(params) == 1 && params[0].Name == "" && params[0].Value == "" {
		return true
	}
	return false
}

func isValidParamsNames(params []constants.Params, names []string) bool {
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

func getCommandInfo(commandID int) (CommandInfo, error) {
	commandInfo, exists := CommandsHandlers[commandID]
	if !exists {
		// The commandID is not valid.
		return CommandInfo{}, fmt.Errorf("invalid commandID")
	}

	// Check if the number of arguments matches the expected count.
	return commandInfo, nil
}

func dissconectPlayer(player *models.Player) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//send response
	err := network.ProcessSendResponseServerErrorDisconnected(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func cubesCanBePlayed(values []int) bool {
	for _, value := range values {
		for _, score := range constants.CGScoreCubeValues {
			if value == score.Value {
				return true
			}
		}
	}
	return false
}

//endregion

//region PROCESS FUNCTIONS

func processResponseClientSucess(player *models.Player, timeStamp string) error {
	err := command_processing_utils.ProcessResponseClientSucessByPlayer(player, timeStamp)
	if err != nil {
		//disconnect player
		logger.Log.Errorf("Error processing response success: %v", err)
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
	}
	return nil
}

func processPlayerLogin(playerNickname string, connectionInfo models.ConnectionInfo, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	//Check if playerNickname in list
	if models.GetInstancePlayerList().HasItemName(playerNickname) {
		err := network.ProcessSendResponseServerErrDuplicitNickname(responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}
	// Add the player to the playerData
	player := models.CreatePlayer(playerNickname, connectionInfo)

	err := models.GetInstancePlayerList().AddItem(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error adding player: %w", err)
	}

	//Send response
	err = sendResponseServerGameList(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Move state machine to lobby
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

//region func processClientCreateGame

func processClientCreateGame(player *models.Player, params []constants.Params, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//State machine check
	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	//Convert params
	gameName, maxPlayers, err := parser.ConvertParamClientCreateGame(params, command.ParamsNames)
	if err != nil {
		logger.Log.Errorf("Error converting params: %v", err)
		errDissconnect := dissconectPlayer(player)
		if errDissconnect != nil {
			errorHandeling.PrintError(errDissconnect)
			return fmt.Errorf("Error sending response: %w", errDissconnect)
		}

		errorHandeling.PrintError(err)
		return nil
	}

	//Check if playerNickname in list
	if models.GetInstanceGameList().HasItemName(gameName) {
		err := network.ProcessSendResponseServerErrorDuplicitGameName(responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Initialize the game
	game, err := initGame(player, gameName, maxPlayers)
	if err != nil {
		return err
	}

	// Send the response
	err = network.SendResponseServerSuccess(responseInfo)
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

	//region ServerUpdateGameList
	err = network.ProcessCommunicationServerUpdateGameList()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//ServerUpdatePlayerList

	err = network.ProcessCommunicationServerUpdatePlayerList(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func sendResponseServerGameList(player *models.Player) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}
	gameList := models.GetInstanceGameList().GetCreatedGameList()

	err := network.SendResponseServerGameList(responseInfo, gameList)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func initGame(player *models.Player, name string, maxPlayers int) (game *models.Game, error error) {
	// Create the game
	game, err := models.CreateGame(name, maxPlayers)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, nil
	}

	// Add the player to the game
	err = game.AddPlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, nil
	}

	_, err = models.GetInstanceGameList().AddItem(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, nil
	}

	return game, nil
}

func _handleCannotFire(player *models.Player) error {

	err := fmt.Errorf("state machine cannot fire")

	logger.Log.Errorf("TOTAL_DISCONNECT: Cannot fire state machine: %v", err)

	network.ImidiateDisconnectPlayer(player.GetNickname())
	return nil
}

//endregion

//region func processClientJoinGame

func processClientJoinGame(player *models.Player, params []constants.Params, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	// Convert params
	gameName, err := parser.ConvertParamClientJoinGame(params, command.ParamsNames) // assuming gameName is the first parameter
	if err != nil {
		logger.Log.Errorf("Error converting params: %v", err)
		errDissconnect := dissconectPlayer(player)
		if errDissconnect != nil {
			errorHandeling.PrintError(errDissconnect)
			return fmt.Errorf("Error sending response: %w", errDissconnect)
		}

		errorHandeling.PrintError(err)
		return nil
	}

	game, err := models.GetInstanceGameList().GetItemByName(gameName)
	if err != nil {
		errorHandeling.PrintError(err)
		logger.Log.Errorf("Error getting game: %v", err)
		errDisconnect := dissconectPlayer(player)
		if errDisconnect != nil {
			errorHandeling.PrintError(errDisconnect)
			return fmt.Errorf("Error sending response: %w", errDisconnect)
		}

		return fmt.Errorf("Error sending response: %w", err)
	}

	err = processAddPlayerToGame(player, game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Send the response
	err = network.SendResponseServerSuccess(responseInfo)
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

	//region ServerUpdateGameList
	err = network.ProcessCommunicationServerUpdateGameList()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	// Send update to all players
	err = network.ProcessCommunicationServerUpdatePlayerList(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func processAddPlayerToGame(player *models.Player, game *models.Game) error {
	var err error

	// Check if player is already in game
	isPlayerInGame := models.GetInstanceGameList().GetPlayersGame(player) != nil
	if isPlayerInGame {
		err = fmt.Errorf("Player is already in a game")
		errorHandeling.PrintError(err)
		return err
	}

	// Check if the game has already started
	canStartGame := game.GetState() == models.Created
	if !canStartGame {
		err = fmt.Errorf("Game has already started")
		errorHandeling.PrintError(err)
		return err
	}

	// Add the player to the game
	err = game.AddPlayer(player)
	if err != nil {
		err = fmt.Errorf("Error adding player to game: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

//endregion

func processClientStartGame(player *models.Player, params []constants.Params, command constants.Command) error {

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//State machine check
	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join playersGame %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	playersGame := models.GetInstanceGameList().GetPlayersGame(player)
	if playersGame == nil {
		logger.Log.Errorf("Error getting playersGame: %v", err)
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//region LOGIC

	//region SendResponseServerSuccess

	err = network.SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//endregion

	//region CommmunicationServerUpdateStartGame

	playersList := playersGame.GetPlayers()
	// remove player from list
	playersList = helpers.RemovePlayerFromList(playersList, player)

	err = network.CommunicationServerUpdateStartGame(playersList)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//endregion

	//region StartGame
	err = playersGame.StartGame()
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
	//endregion

	//region ServerUpdateGameList
	err = network.ProcessCommunicationServerUpdateGameList()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//region ServerUpdateGameData
	err = network.ProcessCommunicationServerUpdateGameData(playersGame)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//endregion

	return nil
}

func processClientPlayerLogout(player *models.Player, params []constants.Params, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	// Convert params
	if len(params) != 0 {
		logger.Log.Errorf("Error number of params: %v", err)
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Send the response
	err = network.SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//Remove from game
	err = models.GetInstanceGameList().RemovePlayerFromGame(player)
	if err != nil {
		err = fmt.Errorf("Error removing player from game: %w", err)
		errorHandeling.PrintError(err)
		panic(err)
	}

	logger.Log.Errorf("Logout player: %v", player.GetNickname())
	//disconnect player
	err = dissconectPlayer(player)
	if err != nil {
		err = fmt.Errorf("Error disconnecting player: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func __handleErrorMyTurn(player *models.Player, game *models.Game) error {
	err := dissconectPlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//next Player turn
	err = game.NextPlayerTurn()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	err = network.ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func validatePlayerTurn(player *models.Player) (*models.Game, error) {
	game := models.GetInstanceGameList().GetPlayersGame(player)
	if game == nil {
		logger.Log.Errorf("Error getting playersGame")
		err := dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return nil, fmt.Errorf("Error sending response: %w", err)
		}
		return nil, nil
	}

	turnPlayer, err := game.GetTurnPlayer()
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, fmt.Errorf("Error sending response: %w", err)
	}

	if turnPlayer.GetNickname() != player.GetNickname() {
		logger.Log.Errorf("Error player is not in turn")
		err := dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return nil, fmt.Errorf("Error sending response: %w", err)
		}
		return nil, nil
	}

	return game, nil
}

func processClientRollDice(player *models.Player, params []constants.Params, command constants.Command) error {
	//inline function _handleCannotFire

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//region CHECK
	//State machine check
	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join playersGame %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	//region check if it is players turn
	game, err := validatePlayerTurn(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	if game == nil {
		return nil
	}
	//endregion
	//endregion

	err = player.GetStateMachine().Fire(command.Trigger)
	if err != nil {
		err = fmt.Errorf("Error firing state machine: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	//region Fork_my_turn
	cubeValues, err := game.NewThrow(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	logger.Log.Debugf("Cube values: %v", cubeValues)

	canBePlayed := cubesCanBePlayed(cubeValues)
	logger.Log.Debugf("Can be played: %v", canBePlayed)
	// endregion

	//region Fork_my_turn -> end 1. ResponseServerEndTurn
	if !canBePlayed {
		//--> ResponseServerEndTurn
		commandTrigger := constants.CGCommands.ResponseServerEndTurn.Trigger
		canFire, err = player.GetStateMachine().CanFire(commandTrigger)
		if err != nil {
			err = fmt.Errorf("Error cannot fire: %w", err)
			errorHandeling.PrintError(err)
			return err
		}
		if !canFire {
			logger.Log.Errorf("Cannot fire with trigger: %v", commandTrigger)
			return __handleErrorMyTurn(player, game)
		}

		err = network.SendResponseServerEndTurn(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = game.NextPlayerTurn()
		if err != nil {
			err = fmt.Errorf("Error next player turn: %w", err)
			errorHandeling.PrintError(err)
			return err
		}

		err = player.FireStateMachine(commandTrigger)
		if err != nil {
			err = fmt.Errorf("Error firing state machine: %w", err)
			errorHandeling.PrintError(err)
			return err
		}

		//ServerUpdateGameData
		err = network.ProcessCommunicationServerUpdateGameData(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		return nil
	}
	//endregion

	//region Fork_my_turn 2. ResponseServerSelectCubes
	commandTrigger := constants.CGCommands.ResponseServerSelectCubes.Trigger
	canFire, err = player.GetStateMachine().CanFire(commandTrigger)
	if err != nil {
		err = fmt.Errorf("Error cannot fire: %w", err)
		errorHandeling.PrintError(err)
		return err
	}
	if !canFire {
		logger.Log.Errorf("Cannot fire with trigger: %v", commandTrigger)
		return __handleErrorMyTurn(player, game)
	}

	err = network.SendResponseServerSelectCubes(cubeValues, responseInfo)
	if err != nil {
		err = fmt.Errorf("Error sending response: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	err = player.FireStateMachine(commandTrigger)
	if err != nil {
		err = fmt.Errorf("Error firing state machine: %w", err)
		errorHandeling.PrintError(err)
		return err
	}
	//endregion

	//region ServerUpdateGameData
	err = network.ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	return nil

	//endregion
}

func processClientSelectedCubes(player *models.Player, params []constants.Params, command constants.Command) error {

	//region CHECK
	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join playersGame %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	//region check if it is players turn
	game, err := validatePlayerTurn(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	if game == nil {
		return nil
	}
	//endregion

	//endregion

	err = player.GetStateMachine().Fire(command.Trigger)
	if err != nil {
		err = fmt.Errorf("Error firing state machine: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	//region Fork_next_dice
	selectedCubesValues, err := parser.ConvertParamClientSelectedCubes(params, command.ParamsNames)
	if err != nil {
		logger.Log.Errorf("Error converting params: %v", err)
		return __handleErrorMyTurn(player, game)
	}

	score, err := game.GetNewScore(player, selectedCubesValues)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//region Fork_next_dice -> end 1. ResponseServerEndScore
	if score >= constants.CMaxScore {

		//check if player can fire
		commandTrigger := constants.CGCommands.ResponseServerEndScore.Trigger
		canFire, err = player.GetStateMachine().CanFire(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			logger.Log.Errorf("Cannot fire with trigger: %v", commandTrigger)
			return __handleErrorMyTurn(player, game)
		}

		//Send ResponseServerEndScore
		err = network.SendResponseServerEndScore(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		//change player score
		err = game.SetPlayerScore(player, selectedCubesValues)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		//Send ServerUpdateEndScore
		sendPlayerList := game.GetPlayers()
		sendPlayerList = helpers.RemovePlayerFromList(sendPlayerList, player)
		err = network.CommunicationServerUpdateEndScore(sendPlayerList, player.GetNickname())
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		//remove game
		err = models.GetInstanceGameList().RemoveItem(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending response: %w", err)
		}

		// set player response succes client 0 for each player
		//for _, p := range game.GetPlayers() {
		//	p.resetResponseSuccessExpected()
		//}

		// fire state machine
		err = player.FireStateMachine(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		// ServerUpdateGameList
		err = network.ProcessCommunicationServerUpdateGameList()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		return nil
	}
	//endregion

	//region Fork_next_dice -> 2. ResponseServerDiceSuccess

	//check if player can fire
	commandTrigger := constants.CGCommands.ResponseServerDiceSuccess.Trigger
	canFire, err = player.GetStateMachine().CanFire(commandTrigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		logger.Log.Errorf("Cannot fire with trigger: %v", commandTrigger)
		return __handleErrorMyTurn(player, game)
	}

	err = game.SetPlayerScore(player, selectedCubesValues)

	err = network.SendResponseServerDiceSuccess(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	err = player.FireStateMachine(commandTrigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//region ServerUpdateGameData
	err = network.ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	return nil
}

func processClientReconnect(player *models.Player, params []constants.Params, command constants.Command) error {
	__disconnectPlayer := func(player *models.Player) error {
		err := dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	inner_send_respones_game_list := func(player *models.Player) error {

		//region SendResponseServerGameList
		logger.Log.Debugf("Processing SendResponseServerGameList: %v", player.GetNickname())
		commandTrigger := constants.CGCommands.ResponseServerGameList.Trigger
		stateMachine := player.GetStateMachine()

		canFire, err := stateMachine.CanFire(commandTrigger)
		if err != nil {
			logger.Log.Errorf("Error cannot fire: %v", err)
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			logger.Log.Errorf("Cannot fire with trigger: %v", commandTrigger)
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}

		err = sendResponseServerGameList(player)
		if err != nil {
			logger.Log.Errorf("Error sending response: %v", err)
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = stateMachine.Fire(commandTrigger)
		if err != nil {
			logger.Log.Errorf("Error firing state machine: %v", err)
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}
		//endregion

		return nil
	}

	logger.Log.Debugf("Processing client reconnect: %v", player.GetNickname())

	//region CHECK
	playerFromList, err := models.GetInstancePlayerList().GetItem(player.GetNickname())
	if err != nil {
		logger.Log.Errorf("Error getting player from list: %v", player.GetNickname())
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	if playerFromList == nil {

		logger.Log.Errorf("Player not found in list: %v", player.GetNickname())
		return __disconnectPlayer(player)
	}

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//region LOGIC
	player.SetConnectedByBool(true)
	player.NullifyTotalDisconnectTime()

	currentStateName := player.GetCurrentStateName()

	// State: Start -> ClientReconnect -> ...
	player.ResetStateMachine()
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		logger.Log.Errorf("Error firing state machine: %v", err)
		errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
	}
	stateMachine := player.GetStateMachine()

	beforeGameAllowedStates := []string{state_machine.StateNameMap.StateGame}
	runningGameAllowedStates := []string{
		state_machine.StateNameMap.StateRunningGame,
		state_machine.StateNameMap.StateMyTurn,
		state_machine.StateNameMap.StateForkMyTurn,
		state_machine.StateNameMap.StateNextDice,
		state_machine.StateNameMap.StateForkNextDice,
	}

	if helpers.Contains(beforeGameAllowedStates, currentStateName) {
		//region RespondServerReconnectBeforeGame
		logger.Log.Debugf("Processing RespondServerReconnectBeforeGame: %v", player.GetNickname())

		game := models.GetInstanceGameList().GetPlayersGame(player)

		if game == nil {
			return inner_send_respones_game_list(player)
		}

		if game.GetState() != models.Created {
			return inner_send_respones_game_list(player)
		}

		commandTrigger := constants.CGCommands.ResponseServerReconnectBeforeGame.Trigger
		canFire, err := stateMachine.CanFire(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}

		err = network.SendResponseServerReconnectBeforeGame(responseInfo, game)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = stateMachine.Fire(commandTrigger)
		if err != nil {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}
		//endregion

		//region CommmunicationServerUpdatePlayerList
		err = network.ProcessCommunicationServerUpdatePlayerList(models.GetInstanceGameList().GetPlayersGame(player))
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		//endregion
		return nil
	}

	if helpers.Contains(runningGameAllowedStates, currentStateName) {
		//region RespondServerReconnectRunningGame
		logger.Log.Debugf("Processing RespondServerReconnectRunningGame: %v", player.GetNickname())

		game := models.GetInstanceGameList().GetPlayersGame(player)
		if game == nil || game.GetState() != models.Running {
			return inner_send_respones_game_list(player)
		}

		commandTrigger := constants.CGCommands.ResponseServerReconnectRunningGame.Trigger
		canFire, err := stateMachine.CanFire(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}

		gameData, err := models.GetInstanceGameList().GetPlayersGame(player).GetGameData()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = network.SendResponseServerReconnectRunningGame(responseInfo, gameData)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = stateMachine.Fire(commandTrigger)
		if err != nil {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}
		//endregion

		//region CommmunicationServerUpdateGameData
		err = network.ProcessCommunicationServerUpdateGameData(models.GetInstanceGameList().GetPlayersGame(player))
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		//endregion

		//region ServerStartTurn - if player is in my turn
		turnPlayer, err := game.GetTurnPlayer()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		if turnPlayer.GetNickname() != player.GetNickname() {
			logger.Log.Debugf("Player %v is not in my turn it is turn Player: %v", player.GetNickname(), turnPlayer.GetNickname())
			return nil
		}

		logger.Log.Debugf("Player %v is in my turn it is turn Player: %v", player.GetNickname(), turnPlayer.GetNickname())

		err = ProcessPlayerTurn(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		//endregion

		return nil
	}

	//endregion

	//endregion
	return inner_send_respones_game_list(player)
}

func ProcessSendPingPlayer(player *models.Player) error {
	commandTrigger := constants.CGCommands.ServerPingPlayer.Trigger
	canFire, err := player.GetStateMachine().CanFire(commandTrigger)
	if err != nil {
		errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
	}
	if !canFire {
		errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
	}

	err = network.CommunicationServerPingPlayer(player)
	if err != nil {
		err = fmt.Errorf("Error pinging player: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}


// region func ProcessPlayerTurn
func handleServerStartTurn(turnPlayer *models.Player) (bool, error) {
	command := constants.CGCommands.ServerStartTurn
	commandTrigger := command.Trigger
	canFire, err := turnPlayer.GetStateMachine().CanFire(commandTrigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("cannot join game %w", err)
	}
	if !canFire {
		err = fmt.Errorf("state machine cannot fire")
		errorHandeling.PrintError(err)
		//is next player turn
		return true, nil
	}

	err = network.CommunicationServerStartTurn(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	err = turnPlayer.FireStateMachine(commandTrigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	return false, nil
}

func ProcessPlayerTurn(game *models.Game) error {
	turnPlayer, err := game.GetTurnPlayer()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//region ServerStartTurn
	isNextPlayerTurn, err := handleServerStartTurn(turnPlayer)
	if err != nil {
		return fmt.Errorf("Error sending response: %w", err)
	}
	if isNextPlayerTurn {
		logger.Log.Errorf("Next player turn")
		return __handleErrorMyTurn(turnPlayer, game)
	}
	//endregion

	//region ServerUpdateGameData
	err = network.ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	//todo remove
	time.Sleep(2 * time.Second)

	//region ServerUpdateGameData
	err = network.ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	return nil
}


//endregion

//endregion

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
)

// CommandsHandlers is a map of valid commands and their information.
var CommandsHandlers = map[int]CommandInfo{
	//1: {"PLAYER_LOGIN", processPlayerLogin} first case in if in ProcessMessage
	constants.CGCommands.ClientCreateGame.CommandID:      {processClientCreateGame, constants.CGCommands.ClientCreateGame},
	constants.CGCommands.ClientJoinGame.CommandID:        {processClientJoinGame, constants.CGCommands.ClientJoinGame},
	constants.CGCommands.ClientStartGame.CommandID:       {processClientStartGame, constants.CGCommands.ClientStartGame},
	constants.CGCommands.ClientLogout.CommandID:          {processClientPlayerLogout, constants.CGCommands.ClientLogout},
	constants.CGCommands.ClientReconnect.CommandID:       {processClientReconnect, constants.CGCommands.ClientReconnect},
	constants.CGCommands.ResponseClientSuccess.CommandID: {processResponseClientSucess, constants.CGCommands.ResponseClientSuccess},
	constants.CGCommands.ClientRollDice.CommandID:        {processClientRollDice, constants.CGCommands.ClientRollDice},
	constants.CGCommands.ClientSelectedCubes.CommandID:   {processClientSelectedCubes, constants.CGCommands.ClientSelectedCubes},

	// ClientRollDice
	// ClientSelectedCubes
	// ClientEndTurn
	//todo probably missing some commands which can come from client
}

//endregion

//region STRUCTURES

// CommandInfo represents information about a command.
type CommandInfo struct {
	Handler func(player *models.Player, params []constants.Params, command constants.Command) error //todo should all return network message format
	Command constants.Command
}

//endregion

func ProcessMessage(message models.Message, conn net.Conn) error {
	logger.Log.Debugf("Starting to process: %v", message)

	// nested function handle invalid message format
	handleInvalidMessageFormat := func(player *models.Player, responseInfo models.MessageInfo) error {
		err := dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending response: %w", err)
		}
		errorHandeling.PrintError(fmt.Errorf("invalid command or incorrect number of arguments"))
		//todo remove
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
		errorHandeling.PrintError(err)
		return fmt.Errorf("invalid command or incorrect number of arguments")
	}
	if player == nil {
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

func processResponseClientSucess(player *models.Player, params []constants.Params, command constants.Command) error {
	return command_processing_utils.ProcessResponseClientSucessByPlayer(player)
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

	errDissconnect := dissconectPlayer(player)
	if errDissconnect != nil {
		errorHandeling.PrintError(errDissconnect)
		return fmt.Errorf("Error sending response: %w", errDissconnect)
	}

	errorHandeling.PrintError(err)
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

	//region CHECK
	playerFromList, err := models.GetInstancePlayerList().GetItem(player.GetNickname())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	if playerFromList == nil {
		return __disconnectPlayer(player)
	}
	if playerFromList.IsConnected() {
		return __disconnectPlayer(player)
	}
	// Convert params
	if len(params) != 0 {
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}
	//endregion

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	//region LOGIC
	player.SetConnectedByBool(true)

	currentStateName := player.GetCurrentStateName()

	// State: Start -> ClientReconnect -> ...
	player.ResetStateMachine()
	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
	}
	stateMachine := player.GetStateMachine()

	//if game ended go to lobby
	game := models.GetInstanceGameList().GetPlayersGame(player)
	if game == nil {
		//region SendResponseServerGameList
		commandTrigger := constants.CGCommands.ResponseServerGameList.Trigger

		canFire, err := stateMachine.CanFire(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}

		err = sendResponseServerGameList(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}

		err = stateMachine.Fire(commandTrigger)
		if err != nil {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}
		//endregion

		return nil
	}

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
		commandTrigger := constants.CGCommands.ResponseServerReconnectBeforeGame.Trigger
		canFire, err := stateMachine.CanFire(commandTrigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("cannot join game %w", err)
		}
		if !canFire {
			errorHandeling.AssertError(fmt.Errorf("cannot fire state machine"))
		}

		err = network.SendResponseServerReconnectBeforeGame(responseInfo, models.GetInstanceGameList().GetValuesArray())
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
		return nil
	}

	return __disconnectPlayer(player)

	//endregion

	//endregion
}

func ProcessSendPingPlayer(player *models.Player) error {
	commandTrigger := constants.CGCommands.ServerPingPlayer.Trigger
	canFire, err := player.GetStateMachine().CanFire(commandTrigger)
	if err != nil {
		err = fmt.Errorf("cannot join game %w", err)
		errorHandeling.PrintError(err)
		return err
	}
	if !canFire {
		err = fmt.Errorf("state machine cannot fire")
		errorHandeling.PrintError(err)
		return err
	}

	err = network.CommunicationServerPingPlayer(player)
	if err != nil {
		err = fmt.Errorf("Error pinging player: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	//fire
	err = player.FireStateMachine(commandTrigger)
	if err != nil {
		err = fmt.Errorf("Error firing state machine: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

//func processClientReconnect(player *models.Player, params []constants.Params, command constants.Command) error {
//	//responseInfo := models.MessageInfo{
//	//	ConnectionInfo: player.GetConnectionInfo(),
//	//	PlayerNickname: player.GetNickname(),
//	//}
//
//	playersGame := models.GetInstanceGameList().GetPlayersGame(player)
//	if playersGame == nil {
//		logger.Log.Errorf("Player %s is not in a game", player.GetNickname())
//		panic("Player is not in a game")
//	}
//
//	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return fmt.Errorf("cannot join game %w", err)
//	}
//	if !canFire {
//		return _handleCannotFire(player)
//	}
//
//	// Convert params
//	if len(params) != 0 {
//		err = dissconectPlayer(player)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		return nil
//	}
//
//	//region LOGIC
//
//	wasHandled := false
//	//ReconnectGameList
//	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ResponseServerReconnectBeforeGame.Trigger)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return fmt.Errorf("cannot join game %w", err)
//	}
//	if canFire {
//		_, err := network.CommunicationServerReconnectGameList(player, models.GetInstanceGameList().GetValuesArray())
//
//		err = player.FireStateMachine(constants.CGCommands.ResponseServerReconnectBeforeGame.Trigger)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		wasHandled = true
//	}
//
//	//ReconnectGameData
//	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ResponseServerReconnectRunningGame.Trigger)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return fmt.Errorf("cannot join game %w", err)
//	}
//	if canFire && !wasHandled {
//		gameData, err := playersGame.GetGameData()
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		_, err = network.CommunicationServerReconnectGameData(player, gameData)
//
//		err = player.FireStateMachine(constants.CGCommands.ResponseServerReconnectRunningGame.Trigger)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		wasHandled = true
//	}
//
//	//ReconnectPlayerList
//	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ServerReconnectPlayerList.Trigger)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return fmt.Errorf("cannot join game %w", err)
//	}
//
//	if canFire && !wasHandled {
//		_, err := network.CommunicationServerReconnectPlayerList(player, playersGame.GetPlayers())
//
//		err = player.FireStateMachine(constants.CGCommands.ServerReconnectPlayerList.Trigger)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		wasHandled = true
//	}
//
//	if !wasHandled {
//		err = dissconectPlayer(player)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return fmt.Errorf("Error sending response: %w", err)
//		}
//		return nil
//	}
//
//	//endregion
//
//	return nil
//}

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

// func _handleServerStartTurn(turnPlayer *models.Player) (bool, error)
// returns:
// - bool: isNextPlayerTurn
// - error: error
//func _handleClientRollDice(turnPlayer *models.Player) (models.Message, bool, error) {
//	command := constants.CGCommands.ClientRollDice
//	message, isSuccess, err := network.CommunicationReadContinouslyWithPing(turnPlayer, command)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return message, false, fmt.Errorf("Error sending response: %w", err)
//	}
//	if !isSuccess {
//		// next Player turn
//		return message, true, nil
//	}
//
//	err = turnPlayer.FireStateMachine(constants.CGCommands.ClientRollDice.Trigger)
//	if err != nil {
//		isSuccess, err = __handleFireError(turnPlayer, err)
//		return message, isSuccess, err
//	}
//
//	return message, false, nil
//}

//func __handleFireError(player *models.Player, err error) (bool, error) {
//	err = fmt.Errorf("Error sending response: %w", err)
//	errorHandeling.PrintError(err)
//
//	errCannotFire := _handleCannotFire(player)
//	return true, errCannotFire
//}

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

	return nil
}

//func ProcessPlayerTurn(game *models.Game) (bool, error) {
//
//	turnPlayer, err := game.GetTurnPlayer()
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return false, fmt.Errorf("Error sending response: %w", err)
//	}
//
//	//region ServerStartTurn
//	isNextPlayerTurn, err := handleServerStartTurn(turnPlayer)
//	if err != nil {
//		return false, fmt.Errorf("Error sending response: %w", err)
//	}
//	if isNextPlayerTurn {
//		return true, nil
//	}
//	//endregion
//
//	//region Start turn
//	for {
//		roundNumber := game.GetRoundNum()
//		logger.Log.Debugf("Turn number: %d ,player: %s ", roundNumber, turnPlayer.GetNickname())
//
//		//region ServerUpdateGameData
//		err = ProcessCommunicationServerUpdateGameData(game, nil)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		//endregion
//
//		//region ClientRollDice
//		var message models.Message
//		message, isNextPlayerTurn, err = _handleClientRollDice(turnPlayer)
//		if err != nil {
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		if isNextPlayerTurn {
//			return true, nil
//		}
//		//endregion
//
//		//region Fork_my_turn
//		cubeValues, err := game.NewThrow(turnPlayer)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		logger.Log.Debugf("Cube values: %v", cubeValues)
//
//		canBePlayed := cubesCanBePlayed(cubeValues)
//		logger.Log.Debugf("Can be played: %v", canBePlayed)
//		// endregion
//
//		//region Fork_my_turn -> end 1. ResponseServerEndTurn
//		if !canBePlayed {
//			//--> ResponseServerEndTurn
//			err = network.SendResponseServerEndTurn(turnPlayer)
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerEndTurn.Trigger)
//			if err != nil {
//				return __handleFireError(turnPlayer, err)
//			}
//
//			//ServerUpdateGameData
//			err = ProcessCommunicationServerUpdateGameData(game, turnPlayer)
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			// next Player turn
//			return true, nil
//		}
//		//endregion
//
//		//region Fork_my_turn 2. ResponseServerSelectCubes
//		err = network.SendResponseServerSelectCubes(cubeValues, message)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//
//		err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerSelectCubes.Trigger)
//		if err != nil {
//			return __handleFireError(turnPlayer, err)
//		}
//		//endregion
//
//		//region ServerUpdateGameData
//		err = ProcessCommunicationServerUpdateGameData(game, nil)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		//endregion
//
//		//region ClientSelectedCubes
//
//		var isSuccess bool
//		command := constants.CGCommands.ClientSelectedCubes
//		message, isSuccess, err = network.CommunicationReadContinouslyWithPing(turnPlayer, command)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		if !isSuccess {
//			//next Player turn
//			return true, nil
//		}
//
//		if message.CommandID != constants.CGCommands.ClientSelectedCubes.CommandID {
//			return false, fmt.Errorf("Error sending response")
//		}
//
//		err = turnPlayer.FireStateMachine(constants.CGCommands.ClientSelectedCubes.Trigger)
//		if err != nil {
//			return __handleFireError(turnPlayer, err)
//		}
//
//		//endregion
//
//		//region Fork_next_dice
//		selectedCubesValues, err := parser.ConvertParamClientSelectedCubes(message.Parameters, constants.CGCommands.ClientSelectedCubes.ParamsNames)
//		if err != nil {
//			errDissconnect := dissconectPlayer(turnPlayer)
//			if errDissconnect != nil {
//				errorHandeling.PrintError(errDissconnect)
//				return false, fmt.Errorf("Error sending response: %w", errDissconnect)
//			}
//
//			errorHandeling.PrintError(err)
//			return true, nil
//		}
//
//		score, err := game.GetNewScore(turnPlayer, selectedCubesValues)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		//endregion
//
//		//region Fork_next_dice -> end 1. ResponseServerEndScore
//		if score >= constants.CMaxScore {
//
//			err = network.SendResponseServerEndScore(turnPlayer)
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			//change player score
//			err = game.SetPlayerScore(turnPlayer, selectedCubesValues)
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			//Send ServerUpdateEndScore
//			playerList := game.GetPlayers()
//			playerList = helpers.RemovePlayerFromList(playerList, turnPlayer)
//
//			err = network.CommunicationServerUpdateEndScore(playerList, turnPlayer.GetNickname())
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerEndScore.Trigger)
//			if err != nil {
//				return __handleFireError(turnPlayer, err)
//			}
//
//			//LOGIC
//			err = models.GetInstanceGameList().RemoveItem(game)
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("error sending response: %w", err)
//			}
//
//			// set player response succes client 0 for each player
//			//for _, p := range game.GetPlayers() {
//			//	p.resetResponseSuccessExpected()
//			//}
//
//			// ServerUpdateGameList
//			err = network.ProcessCommunicationServerUpdateGameList(models.GetInstancePlayerList().GetValuesArray(), models.GetInstanceGameList().GetValuesArray())
//			if err != nil {
//				errorHandeling.PrintError(err)
//				return false, fmt.Errorf("Error sending response: %w", err)
//			}
//
//			return false, nil
//		}
//		//endregion
//
//		//region Fork_next_dice -> 2. ResponseServerDiceSuccess
//		err = game.SetPlayerScore(turnPlayer, selectedCubesValues)
//
//		err = network.SendResponseServerDiceSuccess(turnPlayer)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//
//		err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerDiceSuccess.Trigger)
//		if err != nil {
//			return __handleFireError(turnPlayer, err)
//		}
//		//endregion
//
//		//region ServerUpdateGameData
//		err = ProcessCommunicationServerUpdateGameData(game, turnPlayer)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return false, fmt.Errorf("Error sending response: %w", err)
//		}
//		//endregion
//	}
//	//endregion
//
//	// next Player turn
//	return true, nil
//}

//endregion

//endregion

package command_processing

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/network"
	"gameserver/internal/parser"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/internal/utils/helpers"
	"net"
)

// region GLOBAL VARIABLES
const (
	cMaxScore = 10000
)

// CommandsHandlers is a map of valid commands and their information.
var CommandsHandlers = map[int]CommandInfo{
	//1: {"PLAYER_LOGIN", processPlayerLogin} first case in if in ProcessMessage
	constants.CGCommands.ClientCreateGame.CommandID: {processClientCreateGame, constants.CGCommands.ClientCreateGame},
	constants.CGCommands.ClientJoinGame.CommandID:   {processClientJoinGame, constants.CGCommands.ClientJoinGame},
	constants.CGCommands.ClientStartGame.CommandID:  {processClientStartGame, constants.CGCommands.ClientStartGame},
	constants.CGCommands.ClientLogout.CommandID:     {processClientPlayerLogout, constants.CGCommands.ClientLogout},
	constants.CGCommands.ClientReconnect.CommandID:  {processClientReconnect, constants.CGCommands.ClientReconnect},
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

	err := helpers.RemovePlayerFromGame(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	//todo close connection

	//send response
	err = network.SendResponseServerError(responseInfo, fmt.Errorf("You have been disconnected"))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	return nil
}

func sendCurrentServerUpdateGameData(game *models.Game) error {
	gameData, err := game.GetGameData()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	err = network.CommunicationServerUpdateGameData(gameData, game.GetPlayers())
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

func processPlayerLogin(playerNickname string, connectionInfo models.ConnectionInfo, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: connectionInfo,
		PlayerNickname: playerNickname,
	}

	//Check if playerNickname in list
	if models.GetInstancePlayerList().HasItem(playerNickname) {
		err := network.SendResponseServerErrDuplicitNickname(responseInfo)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Send response
	err := network.SendResponseServerGameList(responseInfo, models.GetInstanceGameList().GetValuesArray())
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Add the player to the playerData
	player := models.CreatePlayer(playerNickname, true, connectionInfo)

	err = models.GetInstancePlayerList().AddItem(player)
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

//region func processClientCreateGame

func processClientCreateGame(player *models.Player, params []constants.Params, command constants.Command) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create game %w", err)
	}
	if !canFire {
		return _handleCannotFire(player)
	}

	//Convert params
	name, maxPlayers, err := parser.ConvertParamClientCreateGame(params, command.ParamsNames)
	if err != nil {
		errDissconnect := dissconectPlayer(player)
		if errDissconnect != nil {
			errorHandeling.PrintError(errDissconnect)
			return fmt.Errorf("Error sending response: %w", errDissconnect)
		}

		errorHandeling.PrintError(err)
		return nil
	}

	err = processCreateGame(player, name, maxPlayers)
	if err != nil {
		return err
	}

	// Send the response
	err = network.SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//ServerUpdateGameList
	err = network.CommunicationServerUpdateGameList(models.GetInstancePlayerList().GetValuesArrayWithoutOnePlayer(player), models.GetInstanceGameList().GetValuesArray())
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

	//ServerUpdatePlayerList

	//player list wit only one element player
	playerList := []*models.Player{player}
	err = network.CommunicationServerUpdatePlayerList(playerList)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func processCreateGame(player *models.Player, name string, maxPlayers int) error {
	// Create the game
	game, err := models.CreateGame(name, maxPlayers)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil
	}

	// Add the player to the game
	err = game.AddPlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil
	}

	_, err = models.GetInstanceGameList().AddItem(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil
	}

	return nil
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

	// Send update to all players
	err = network.CommunicationServerUpdatePlayerList(game.GetPlayers())
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

	//region LOGIC

	// Check if player got playersGame
	playersGame := models.GetInstanceGameList().GetPlayersGame(player)
	if playersGame == nil {
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	// Check if the playersGame has already started
	if playersGame.GetState() != models.Created {
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//Check if minimum players
	if !playersGame.IsEnoughPlayers() {
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	playersGame.SetState(models.Running)

	// Send the response
	err = network.SendResponseServerSuccess(responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	// Send update to all players

	//list which is an copy of players in game
	playersList := playersGame.GetPlayers()
	//remove player from list
	playersList = helpers.RemovePlayerFromList(playersList, player)

	err = network.CommunicationServerUpdateStartGame(playersList)
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
	err = sendCurrentServerUpdateGameData(playersGame)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	//region TURN logic
	//Running_Game --ServerStartTurn--> my_turn
	var isNextPlayerTurn = true
	for isNextPlayerTurn {
		isNextPlayerTurn, err = startPlayerTurn(playersGame)
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

	//region LOGIC
	err = helpers.RemovePlayerFromGame(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

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

	return nil
}

func processClientReconnect(player *models.Player, params []constants.Params, command constants.Command) error {
	//responseInfo := models.MessageInfo{
	//	ConnectionInfo: player.GetConnectionInfo(),
	//	PlayerNickname: player.GetNickname(),
	//}

	playersGame := models.GetInstanceGameList().GetPlayersGame(player)

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

	//region LOGIC

	wasHandled := false
	//ReconnectGameList
	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ServerReconnectGameList.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if canFire {
		_, err := network.CommunicationServerReconnectGameList(player, models.GetInstanceGameList().GetValuesArray())

		err = player.FireStateMachine(constants.CGCommands.ServerReconnectGameList.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	//ReconnectGameData
	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ServerReconnectGameData.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}
	if canFire && !wasHandled {
		gameData, err := playersGame.GetGameData()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		_, err = network.CommunicationServerReconnectGameData(player, gameData)

		err = player.FireStateMachine(constants.CGCommands.ServerReconnectGameData.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	//ReconnectPlayerList
	canFire, err = player.GetStateMachine().CanFire(constants.CGCommands.ServerReconnectPlayerList.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot join game %w", err)
	}

	if canFire && !wasHandled {
		_, err := network.CommunicationServerReconnectPlayerList(player, playersGame.GetPlayers())

		err = player.FireStateMachine(constants.CGCommands.ServerReconnectPlayerList.Trigger)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		wasHandled = true
	}

	if !wasHandled {
		err = dissconectPlayer(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("Error sending response: %w", err)
		}
		return nil
	}

	//endregion

	return nil
}

//region func startPlayerTurn

func startPlayerTurn(game *models.Game) (bool, error) {

	//nested function
	__handleFireError := func(player *models.Player, err error) (bool, error) {
		err = fmt.Errorf("Error sending response: %w", err)
		errorHandeling.PrintError(err)

		errCannotFire := _handleCannotFire(player)
		return true, errCannotFire
	}

	gameData, err := game.GetGameData()
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	turnPlayer := gameData.TurnPlayer
	isSuccess, err := network.CommunicationServerStartTurn(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	if !isSuccess {
		//next Player turn
		return true, nil
	}

	err = turnPlayer.FireStateMachine(constants.CGCommands.ServerStartTurn.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	//My turn --ClientRollDice--> fork_my_turn
	message, isSuccess, err := network.CommunicationReadClientRollDice(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	if !isSuccess {
		//next Player turn
		return true, nil
	}

	err = turnPlayer.FireStateMachine(constants.CGCommands.ClientRollDice.Trigger)
	if err != nil {
		return __handleFireError(turnPlayer, err)
	}

	//fork_my_turn --> ResponseServerEndTurn
	//             --> ResponseServerSelectCubes
	cubeValues, err := game.NewThrow(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	canBePlayed := cubesCanBePlayed(cubeValues)
	if !canBePlayed {
		//--> ResponseServerEndTurn
		err = network.SendResponseServerEndTurn(cubeValues, turnPlayer, message)
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

		err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerEndTurn.Trigger)
		if err != nil {
			return __handleFireError(turnPlayer, err)
		}

		// next Player turn
		return true, nil
	}

	//-->ResponseServerSelectCubes
	err = network.SendResponseServerSelectCubes(cubeValues, turnPlayer, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerSelectCubes.Trigger)
	if err != nil {
		return __handleFireError(turnPlayer, err)
	}

	//Next_dice --> Running_Game
	//           --> Fork_next_dice
	message, isSuccess, err = network.CommunicationReadClient(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}
	if !isSuccess {
		//next Player turn
		return true, nil
	}

	if message.CommandID == constants.CGCommands.ClientEndTurn.CommandID {
		//--> Running_Game

		//Send Success
		responseInfo := models.MessageInfo{
			ConnectionInfo: models.ConnectionInfo{
				Connection: turnPlayer.GetConnectionInfo().Connection,
				TimeStamp:  message.TimeStamp,
			},
			PlayerNickname: message.PlayerNickname,
		}
		err = network.SendResponseServerSuccess(responseInfo)
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

		err = turnPlayer.FireStateMachine(constants.CGCommands.ClientEndTurn.Trigger)
		if err != nil {
			return __handleFireError(turnPlayer, err)
		}

		//next Player turn
		return true, nil
	}

	//--> Fork_next_dice
	if message.CommandID != constants.CGCommands.ClientSelectedCubes.CommandID {
		return false, fmt.Errorf("Error sending response")
	}

	err = turnPlayer.FireStateMachine(constants.CGCommands.ClientSelectedCubes.Trigger)
	if err != nil {
		return __handleFireError(turnPlayer, err)
	}

	//Fork_next_dice --> ResponseServerEndScore
	//               --> ResponseServerDiceSuccess
	selectedCubesValues, err := parser.ConvertParamClientSelectedCubes(message.Parameters, constants.CGCommands.ClientSelectedCubes.ParamsNames)
	if err != nil {
		errDissconnect := dissconectPlayer(turnPlayer)
		if errDissconnect != nil {
			errorHandeling.PrintError(errDissconnect)
			return false, fmt.Errorf("Error sending response: %w", errDissconnect)
		}

		errorHandeling.PrintError(err)
		return true, nil
	}

	scoreIncrease, err := game.GetScoreIncrease(selectedCubesValues, turnPlayer)
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

	// --> ResponseServerEndScore
	if score >= cMaxScore {

		err = network.SendResponseServerEndScore(turnPlayer)
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

		//Send ServerUpdateEndScore
		playerList := game.GetPlayers()
		playerList = helpers.RemovePlayerFromList(playerList, turnPlayer)
		err = network.CommunicationServerUpdateEndScore(playerList, turnPlayer.GetNickname())
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("Error sending response: %w", err)
		}

		err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerEndScore.Trigger)
		if err != nil {
			return __handleFireError(turnPlayer, err)
		}

		err = models.GetInstanceGameList().RemoveItem(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("error sending response: %w", err)
		}

		return false, nil
	}

	//--> ResponseServerDiceSuccess
	err = network.SendResponseServerDiceSuccess(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("Error sending response: %w", err)
	}

	err = turnPlayer.FireStateMachine(constants.CGCommands.ResponseServerDiceSuccess.Trigger)
	if err != nil {
		return __handleFireError(turnPlayer, err)
	}

	// next Player turn
	return true, nil
}

//endregion

//endregion
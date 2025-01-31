package network

import (
	"errors"
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/models/state_machine"
	"gameserver/internal/network/network_utils"
	"gameserver/internal/parser"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/internal/utils/helpers"
	"net"
	"strings"
	"time"
)

//endregion

// enum
type ResponseResult int

const (
	Succes ResponseResult = iota
	Timeout
	WrongResponse
	Error
)

// region PRIVATE SHARED WITH - SERVER_LISTEN
func CloseConnection(connection net.Conn) error {
	err := connection.Close()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}
	errorHandeling.PrintError(fmt.Errorf("Connection closed"))

	logger.Log.Info("Connection closed")
	return nil
}

func Read(connection net.Conn) ([]models.Message, bool, error) {
	messageList, isTimeout, err := connectionReadTimeout(connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return []models.Message{}, isTimeout, fmt.Errorf("error reading %w", err)
	}

	return messageList, isTimeout, nil
}

//endregion

// region PRIVATE FUNCTIONS

func sendResponseServerSelectCubes(command constants.Command, values []int, responseInfo models.MessageInfo) error {
	paramsValue := parser.ConvertListCubeValuesToNetworkString(values)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	err = sendMessageWrapper(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func sendMessageWrapper(messageInfo models.MessageInfo, command constants.Command, params []constants.Params) error {
	_, err := sendMessage(messageInfo, command, params)
	return err
}

func sendMessage(messageInfo models.MessageInfo, command constants.Command, params []constants.Params) (models.Message, error) {
	connection := messageInfo.ConnectionInfo.Connection
	playerName := messageInfo.PlayerNickname

	//convert to message
	message := models.CreateMessage(playerName, command.CommandID, params)

	err := connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, fmt.Errorf("error writing %w", err)
	}
	return message, nil
}

// sendMessageWithSuccessResponse
// Return: bool isSuccess - if communication was successful
func sendMessageWithSuccessResponse(player *models.Player, command constants.Command, params []constants.Params) error {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	message, err := sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	//increase expected response
	player.IncreaseResponseSuccessExpected(message)

	//connection := responseInfo.ConnectionInfo.Connection
	//
	////wait for client response
	//
	//clientResponse, isTimeout, err := ReadSingleTimeout(connection)
	//if err != nil {
	//	errorHandeling.PrintError(err)
	//	return false, fmt.Errorf("error reading %w", err)
	//}
	//if isTimeout {
	//	err = DisconnectPlayerConnection(player)
	//	if err != nil {
	//		errorHandeling.PrintError(err)
	//		return false, fmt.Errorf("error handling timeout %w", err)
	//	}
	//	return false, nil
	//}
	//
	////check if client responded with success
	//if !isClientResponseCommand(clientResponse, responseInfo.PlayerNickname, constants.CGCommands.ResponseClientSuccess) {
	//	return false, _handleClienResponseNotSuccess(player)
	//}
	//
	//return true, nil

	return nil
}

func sendUpdateList(player *models.Player, command constants.Command, params []constants.Params) error {

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error checking if can fire %w", err)
	}
	if !canFire {
		return nil
	}

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	var message models.Message
	message, err = sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	// set for player expected ClientResponseSuccess which is handled in server_listen.go
	player.IncreaseResponseSuccessExpected(message)

	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error firing state machine %w", err)
	}

	return nil
}


func DisconnectPlayerConnection(player *models.Player) error {
	//commandError := constants.CGCommands.ErrorPlayerUnreachable
	//err := player.FireStateMachine(commandError.Trigger)
	//if err != nil {
	//	errorHandeling.PrintError(err)
	//	return fmt.Errorf("error firing state machine %w", err)
	//}
	player.SetConnectedByBool(false)

	player.SetTotalDisconnectStartTime()
	go processTotalDisconnect(player)

	//close connection
	err := CloseConnection(player.GetConnectionInfo().Connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}

	//region Send updates to all players
	game := models.GetInstanceGameList().GetPlayersGame(player)
	if game == nil {
		return nil
	}

	err = SendGameUpdates(game)
	if err != nil {
		err = fmt.Errorf("Error sending game updates: %w", err)
		errorHandeling.PrintError(err)
		return err
	}
	//endregion

	//start thread which will wait specifed time and if player is not yet connected again, set that he is disconnected in game

	return nil
}

func totalDisconnect(player *models.Player) {

	logger.Log.Infof("TOTAL_DISCONNECT: Player %s has not reconnected in time", player.GetNickname())

	//player havent reconnected in the wait -> total disconnect
	playerFromList, err := models.GetInstancePlayerList().GetItem(player.GetNickname())
	if err != nil {
		errorHandeling.PrintError(err)
		return
	}
	if playerFromList == nil {
		errorHandeling.AssertError(fmt.Errorf("error player is nil"))
	}

	playerFromList.SetConnected(models.ConnectionStates.TotalDisconnect)

	//remove player from game
	game := models.GetInstanceGameList().GetPlayersGame(player)
	if game == nil {
		return
	}

	err = helpers.RemovePlayerFromLists(player)

	logger.Log.Infof("TOTAL_DISCONNECT: Player %s has not reconnected in time", player.GetNickname())

	//send updates
	err = SendAllUpdates(game)
	if err != nil {
		err = fmt.Errorf("Error sending game updates: %w", err)
		errorHandeling.PrintError(err)
		return
	}

	//players are put back to lobby if only one in runnin game -> others total disconnect
	if !game.IsEnoughPlayersToContinueGame() {
		if game.GetState() != models.Running {
			return
		}

		//ServerUpdateNotEnoughPlayers
		playerList := game.GetPlayers()
		playerList = helpers.PlayerListGetActivePlayers(playerList)
		err = CommunicationServerUpdateNotEnoughPlayers(playerList)
		if err != nil {
			errorHandeling.PrintError(err)
			return
		}

		//remove game
		err = models.GetInstanceGameList().RemoveItem(game)
		if err != nil {
			errorHandeling.PrintError(err)
			return
		}
	}
}

func processTotalDisconnect(player *models.Player) {
	time.Sleep(constants.CTotalDisconnectTime)

	if player.IsTotalDisconnectTimeout() {
		if player.WasTotalDisconnecTimeoutCalled() {
			return
		}

		ImidiateDisconnectPlayer(player.GetNickname())
	}
	//logger.Log.Error("Total disconnect: from timeout")
	//totalDisconnect(player)
}

func connectionReadTimeout(connection net.Conn) ([]models.Message, bool, error) {
	isTimeout := false
	timeout := constants.CTimeout
	// Set the timeout
	deadline := time.Now().Add(timeout)
	err := connection.SetReadDeadline(deadline)
	if err != nil {
		errorHandeling.PrintError(err)
		return []models.Message{}, isTimeout, fmt.Errorf("error setting read deadline: %w", err)
	}

	buffer := make([]byte, constants.CMessageBufferSize)
	messageStr := ""
	maxMessageLength := constants.CMessageMaxSize // Define the maximum messageList length

	for {
		_, err = connection.Read(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				isTimeout = true
				return []models.Message{}, isTimeout, nil
			}
			return []models.Message{}, isTimeout, fmt.Errorf("error reading: %w", err)
		}
		messageStr += string(buffer)

		// Check if the messageList contains the end delimiter
		if strings.Contains(messageStr, constants.CMessageEndDelimiter) {
			break
		}

		//remove whitespace
		trimedMessage := strings.TrimSpace(messageStr)
		// Check if the messageList has reached the maximum length
		if len(trimedMessage) >= maxMessageLength {
			break
		}
	}

	messageList, err := parser.ParseReceiveMessageStr(messageStr)
	if err != nil {
		//if error when reading client message
		logger.Log.Errorf("TOTAL_DISCONNECT: Wrong format message")
		ImidiateDisconnectPlayerByConnection(connection)
	}

	//Save to logger
	for _, message := range messageList {
		network_utils.GReceivedMessageList.AddItem(message)
	}

	return messageList, false, nil
}

func connectionWrite(connection net.Conn, message models.Message) error {
	messageStr, err := parser.ConvertMessageToNetworkString(message)
	if err != nil {
		errorHandeling.AssertError(fmt.Errorf("error converting message to network string"))
	}

	_, err = connection.Write([]byte(messageStr))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing", err)
	}

	network_utils.GSendMessageList.AddItem(message)

	return nil
}


//endregion

// region SEND RESPONSE FUNCTIONS
func sendResponseEmpty(playerName string, connection net.Conn, commandID int) error {
	//convert to message
	message := models.CreateMessage(playerName, commandID, constants.CGNetworkEmptyParams)

	err := connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing %w", err)
	}

	return nil
}

func SendResponseServerSuccess(responseInfo models.MessageInfo) error {
	return sendResponseEmpty(responseInfo.PlayerNickname, responseInfo.ConnectionInfo.Connection, constants.CGCommands.ResponseServerSuccess.CommandID)
}

func ProcessSendResponseServerErrDuplicitNickname(responseInfo models.MessageInfo) error {
	err := processResponseServerError(responseInfo, "error duplicate nickname")
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

// SendResponseServerErrorDupolicitGameName
func ProcessSendResponseServerErrorDuplicitGameName(responseInfo models.MessageInfo) error {
	err := processResponseServerError(responseInfo, "error duplicate game name")
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

// ProcessSendResponseServerErrorDisconnected
func ProcessSendResponseServerErrorDisconnected(responseInfo models.MessageInfo) error {
	err := processResponseServerError(responseInfo, "you have been disconnected")
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func _sendResponseServerError(responseInfo models.MessageInfo, errorStr string) error {
	command := constants.CGCommands.ResponseServerError

	params, err := models.CreateParams(command.ParamsNames, []string{errorStr})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	err = sendMessageWrapper(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func processResponseServerError(responseInfo models.MessageInfo, errorStr string) error {
	err := _sendResponseServerError(responseInfo, errorStr)
	if err != nil {
		err = fmt.Errorf("error sending response %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	//disconnect player
	player, err := models.GetInstancePlayerList().GetItem(responseInfo.PlayerNickname)
	if err != nil {
		//fatal
		err = fmt.Errorf("error getting player %w", err)
		errorHandeling.PrintError(err)
		panic(err)
	}

	if player == nil {
		errorHandeling.AssertError(fmt.Errorf("error player is nil"))
	}

	err = DisconnectPlayerConnection(player)
	if err != nil {
		err = fmt.Errorf("error disconnecting player %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func ImidiateDisconnectPlayer(playerNickname string) {
	player, err := models.GetInstancePlayerList().GetItem(playerNickname)
	if err != nil {
		//fatal
		err = fmt.Errorf("error getting player %w", err)
		errorHandeling.PrintError(err)
		return
	}
	if player == nil {
		errorHandeling.PrintError(fmt.Errorf("error player is nil"))
		return
	}

	player.SetWasTotalDisconnectCalled()

	logger.Log.Errorf("Imidiate Disconnect not from timeout but forces")
	totalDisconnect(player)
}

func ImidiateDisconnectPlayerByConnection(connection net.Conn) {
	player := models.GetInstancePlayerList().GetPlayerByConnection(connection)
	if player == nil {
		err := CloseConnection(connection)
		if err != nil {
			errorHandeling.PrintError(err)
			return
		}
	}
	ImidiateDisconnectPlayer(player.GetNickname())
}

func sendGameList(command constants.Command, responseInfo models.MessageInfo, paramsValues []*models.Game) error {
	value := parser.ConvertListGameListToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error creating params %w", err)
	}
	err = sendMessageWrapper(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func SendResponseServerGameList(responseInfo models.MessageInfo, paramsValues []*models.Game) error {
	command := constants.CGCommands.ResponseServerGameList
	return sendGameList(command, responseInfo, paramsValues)
}

func SendResponseServerReconnectBeforeGame(reponseInfo models.MessageInfo, game *models.Game) error {
	command := constants.CGCommands.ResponseServerReconnectBeforeGame
	messageDataplayersInGameList := game.GetPlayers()

	params, err := prepareParamsForPlayerList(command, messageDataplayersInGameList)
	if err != nil {
		return fmt.Errorf("error creating params %w", err)
	}

	err = sendMessageWrapper(reponseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerReconnectRunningGame(responseInfo models.MessageInfo, paramsValues models.GameData) error {
	command := constants.CGCommands.ResponseServerReconnectRunningGame
	paramsValue := parser.ConvertListGameDataToNetworkString(paramsValues)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	err = sendMessageWrapper(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func SendResponseServerSelectCubes(values []int, responseInfo models.MessageInfo) error {
	command := constants.CGCommands.ResponseServerSelectCubes

	err := sendResponseServerSelectCubes(command, values, responseInfo)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerEndTurn(player *models.Player) error {
	command := constants.CGCommands.ResponseServerEndTurn

	return sendResponseEmpty(player.GetNickname(), player.GetConnectionInfo().Connection, command.CommandID)
}

func SendResponseServerDiceSuccess(player *models.Player) error {
	command := constants.CGCommands.ResponseServerDiceSuccess

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendMessageWrapper(responseInfo, command, constants.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func SendResponseServerEndScore(player *models.Player) error {
	command := constants.CGCommands.ResponseServerEndScore

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendMessageWrapper(responseInfo, command, constants.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

//endregion

//region COMMUNICATION FUNCTIONS

// region SEND FUNCTIONS
func sendStandardUpdateToAllMessage(playerList []*models.Player, command constants.Command, params []constants.Params) error {

	for _, player := range playerList {
		//null responseExpected
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}
	return nil
}

// ProcessCommunicationServerUpdateGameList
// args: playerList - list of players, paramsValues - list of games, player - player that sent the message

func CommunicationServerUpdateGameList(playerList []*models.Player, paramsValues []*models.Game) error {
	command := constants.CGCommands.ServerUpdateGameList
	paramsValue := parser.ConvertListGameListToNetworkString(paramsValues)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	err = sendStandardUpdateToAllMessage(playerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

func prepareParamsForPlayerList(command constants.Command, playersInGameList []*models.Player) ([]constants.Params, error) {
	paramsValue := parser.ConvertListPlayerListToNetworkString(playersInGameList)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, err
	}
	return params, nil
}

func CommunicationServerUpdatePlayerList(sendPlayerList []*models.Player, playersInGameList []*models.Player) error {
	command := constants.CGCommands.ServerUpdatePlayerList

	params, err := prepareParamsForPlayerList(command, playersInGameList)
	if err != nil {
		return fmt.Errorf("error creating params %w", err)
	}

	err = sendStandardUpdateToAllMessage(sendPlayerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

func CommunicationServerUpdateStartGame(playerList []*models.Player) error {
	command := constants.CGCommands.ServerUpdateStartGame
	params := constants.CGNetworkEmptyParams

	err := sendStandardUpdateToAllMessage(playerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

func CommunicationServerUpdateNotEnoughPlayers(playerList []*models.Player) error {
	command := constants.CGCommands.ServerUpdateNotEnoughPlayers
	params := constants.CGNetworkEmptyParams

	err := sendStandardUpdateToAllMessage(playerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

func CommunicationServerUpdateGameData(sendPlayerList []*models.Player, gameData models.GameData) error {
	command := constants.CGCommands.ServerUpdateGameData
	paramsValue := parser.ConvertListGameDataToNetworkString(gameData)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	err = sendStandardUpdateToAllMessage(sendPlayerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

func CommunicationServerUpdateEndScore(playerList []*models.Player, playerName string) error {
	command := constants.CGCommands.ServerUpdateEndScore

	params, err := models.CreateParams(command.ParamsNames, []string{playerName})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	err = sendStandardUpdateToAllMessage(playerList, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending update list %w", err)
	}

	return nil
}

//region SERVER -> SINGLE CLIENT

// CommunicationServerUpdateGameData
func CommunicationServerStartTurn(player *models.Player) error {
	command := constants.CGCommands.ServerStartTurn

	err := sendMessageWithSuccessResponse(player, command, constants.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending ping player %w", err)
	}

	return nil
}

// CommunicationServerPingPlayer
func CommunicationServerPingPlayer(player *models.Player) error {
	command := constants.CGCommands.ServerPingPlayer

	err := sendMessageWithSuccessResponse(player, command, constants.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending ping player %w", err)
	}
	return err
}


//endregion

//endregion

// region RECEIVE FUNCTIONS

//endregion

//endregion

func SendGameUpdates(game *models.Game) error {
	logger.Log.Debugf("SendGameUpdates: Sending Disconnect updates for game")

	//region Send updates to all players
	err := ProcessCommunicationServerUpdatePlayerList(game)
	if err != nil {
		err = fmt.Errorf("Error sending update: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	err = ProcessCommunicationServerUpdateGameData(game)
	if err != nil {
		err = fmt.Errorf("Error sending update: %w", err)
		errorHandeling.PrintError(err)
		return err
	}
	//endregion
	return nil
}

func SendAllUpdates(game *models.Game) error {
	//region ServerUpdateGameList
	err := ProcessCommunicationServerUpdateGameList()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}
	//endregion

	err = SendGameUpdates(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func ProcessCommunicationServerUpdateGameData(game *models.Game) error {

	logger.Log.Infof("ProcessCommunicationServerUpdateGameData")
	if game.GetState() != models.Running {
		return nil
	}

	gameData, err := game.GetGameData()
	if err != nil {
		return nil
	}

	playersInGameList := game.GetPlayers()
	playerList := helpers.PlayerListGetActivePlayers(playersInGameList)

	if len(playerList) == 0 {
		logger.Log.Debugf("ProcessCommunicationServerUpdateGameData: No players in game")
		return nil
	}

	err = CommunicationServerUpdateGameData(playerList, gameData)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func getPlayerListInfo(game *models.Game) ([]*models.Player, []*models.Player) {
	messageDataplayersInGameList := game.GetPlayers()
	sendPlayerList := helpers.PlayerListGetActivePlayersInState(messageDataplayersInGameList, state_machine.StateNameMap.StateGame)

	if len(sendPlayerList) == 0 {
		return nil, nil
	}

	return sendPlayerList, messageDataplayersInGameList
}

func ProcessCommunicationServerUpdatePlayerList(game *models.Game) error {
	logger.Log.Infof("ProcessCommunicationServerUpdatePlayerList")
	playerList, playersInGameList := getPlayerListInfo(game)
	if playerList == nil {
		logger.Log.Debugf("ProcessCommunicationServerUpdatePlayerList: No players in game")
		return nil
	}

	err := CommunicationServerUpdatePlayerList(playerList, playersInGameList)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func ProcessCommunicationServerUpdateGameList() error {
	gameList := models.GetInstanceGameList().GetCreatedGameList()
	playerList := models.GetInstancePlayerList().GetActivePlayersInState(state_machine.StateNameMap.StateLobby)

	if len(playerList) == 0 {
		return nil
	}

	//Send only games which are in state created

	err := CommunicationServerUpdateGameList(playerList, gameList)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

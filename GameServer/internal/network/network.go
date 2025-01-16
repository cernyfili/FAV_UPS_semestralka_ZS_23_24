package network

import (
	"errors"
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
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

// region PRIVATE FUNCTIONS
func CloseConnection(connection net.Conn) error {
	err := connection.Close()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}
	//todo remove
	errorHandeling.PrintError(fmt.Errorf("Connection closed"))

	logger.Log.Info("Connection closed")
	return nil
}

func sendResponseServerSelectCubes(command constants.Command, values []int, player *models.Player, message models.Message) error {
	paramsValue := parser.ConvertListCubeValuesToNetworkString(values)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	responseInfo := models.MessageInfo{
		ConnectionInfo: models.ConnectionInfo{
			Connection: player.GetConnectionInfo().Connection,
			TimeStamp:  message.TimeStamp,
		},
		PlayerNickname: message.PlayerNickname,
	}

	err = sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func sendMessage(messageInfo models.MessageInfo, command constants.Command, params []constants.Params) error { //todo refactor where used
	connection := messageInfo.ConnectionInfo.Connection
	playerName := messageInfo.PlayerNickname

	//convert to message
	message := models.CreateMessage(playerName, command.CommandID, params)

	err := connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing %w", err)
	}
	return nil
}

// sendMessageWithSuccessResponse
// Return: bool isSuccess - if communication was successful
func sendMessageWithSuccessResponse(player *models.Player, command constants.Command, params []constants.Params) (bool, error) {
	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error sending response %w", err)
	}

	connection := responseInfo.ConnectionInfo.Connection

	//wait for client response

	clientResponse, isTimeout, err := ConnectionReadTimeout(connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error reading %w", err)
	}
	if isTimeout {
		err = HandleTimeout(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("error handling timeout %w", err)
		}
		return false, nil
	}

	//check if client responded with success
	if !IsClientResponseCommand(clientResponse, responseInfo.PlayerNickname, constants.CGCommands.ResponseClientSuccess) {
		return false, HandleClienResponseNotSuccess(player)
	}

	return true, nil
}

func HandleClienResponseNotSuccess(player *models.Player) error {
	err := helpers.RemovePlayerFromGame(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error removing player from lists %w", err)
	}

	err = CloseConnection(player.GetConnectionInfo().Connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}
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

	err = sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	// set for player expected ClientResponseSuccess which is handled in server_listen.go
	player.IncreaseResponseSuccessExpected()

	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error firing state machine %w", err)
	}

	return nil
}

func readContinouslyWithPing(player *models.Player, command constants.Command) (models.Message, error) {

	var clientResponse models.Message
	var isTimeout bool
	var err error

	connection := player.GetConnectionInfo().Connection

	for {
		//reading clientRollDice
		clientResponse, isTimeout, err = ConnectionReadTimeout(connection)

		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, fmt.Errorf("error reading %w", err)
		}
		if isTimeout {
			// freeze computer
			//todo remove
			time.Sleep(2 * time.Second)

			responseMessage, isSuccess, err := CommunicationServerPingPlayer(player, command)
			if err != nil {
				errorHandeling.PrintError(err)
				return models.Message{}, fmt.Errorf("error sending ping player %w", err)
			}
			if responseMessage.PlayerNickname != "" {
				//when as response for ping i get an message i want
				return responseMessage, nil
			}
			if !isSuccess {
				//handle timeout
				err := HandleTimeout(player)
				if err != nil {
					errorHandeling.PrintError(err)
					return models.Message{}, fmt.Errorf("error handling timeout %w", err)
				}
				return models.Message{}, err
			}
			continue
		}

		if clientResponse.CommandID == constants.CGCommands.ResponseClientSuccess.CommandID {
			continue
		}

		break
	}
	return clientResponse, nil
}

func communicationRead(player *models.Player, connection net.Conn) (models.Message, error) {

	var clientResponse models.Message
	var isTimeout bool
	var err error
	doesRespond := true
	for doesRespond {
		//reading clientRollDice
		clientResponse, isTimeout, err = ConnectionReadTimeout(connection)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, fmt.Errorf("error reading %w", err)
		}
		if isTimeout {
			err := HandleTimeout(player)
			if err != nil {
				errorHandeling.PrintError(err)
				return models.Message{}, fmt.Errorf("error handling timeout %w", err)
			}
			break
			//
			//isSuccess, err := CommunicationServerPingPlayer(player)
			//if err != nil {
			//	errorHandeling.PrintError(err)
			//	return models.Message{}, models.Message{}, fmt.Errorf("error sending ping player %w", err)
			//}
			//if !isSuccess {
			//	//handle timeout
			//	err := HandleTimeout(player)
			//	if err != nil {
			//		errorHandeling.PrintError(err)
			//		return models.Message{}, models.Message{}, fmt.Errorf("error handling timeout %w", err)
			//	}
			//	break
			//}
		}
	}
	return clientResponse, nil
}

func HandleTimeout(player *models.Player) error {
	commandError := constants.CGCommands.ErrorPlayerUnreachable
	err := player.FireStateMachine(commandError.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error firing state machine %w", err)
	}

	player.SetConnected(false)

	//close connection
	err = CloseConnection(player.GetConnectionInfo().Connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}

	return nil
}

func ConnectionReadTimeout(connection net.Conn) (models.Message, bool, error) {
	isTimeout := false
	timeout := constants.CTimeout
	// Set the timeout
	deadline := time.Now().Add(timeout)
	err := connection.SetReadDeadline(deadline)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, fmt.Errorf("error setting read deadline: %w", err)
	}

	buffer := make([]byte, constants.CMessageBufferSize)
	messageStr := ""
	maxMessageLength := constants.CMessageMaxSize // Define the maximum message length

	for {
		_, err = connection.Read(buffer)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				isTimeout = true
				return models.Message{}, isTimeout, nil
			}
			return models.Message{}, isTimeout, fmt.Errorf("error reading: %w", err)
		}
		messageStr += string(buffer)

		// Check if the message contains the end delimiter
		if strings.Contains(messageStr, constants.CMessageEndDelimiter) {
			break
		}

		//remove whitespace
		trimedMessage := strings.TrimSpace(messageStr)
		// Check if the message has reached the maximum length
		if len(trimedMessage) >= maxMessageLength {
			break
		}
	}

	message, err := parser.ParseMessage(messageStr)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, fmt.Errorf("Error parsing message: %w", err)
	}

	//Save to logger
	err = network_utils.GReceivedMessageList.AddItem(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, err
	}

	return message, false, nil
}

func connectionWrite(connection net.Conn, message models.Message) error {
	messageStr, err := parser.ConvertMessageToNetworkString(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	_, err = connection.Write([]byte(messageStr))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing", err)
	}

	err = network_utils.GSendMessageList.AddItem(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func IsClientResponseCommand(clientResponse models.Message, expactedPlayerName string, command constants.Command) bool {
	if clientResponse.PlayerNickname != expactedPlayerName {
		return false
	}

	if clientResponse.CommandID != command.CommandID {
		return false
	}

	return true
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

func SendResponseServerErrDuplicitNickname(responseInfo models.MessageInfo) error {
	err := SendResponseServerError(responseInfo, fmt.Errorf("error duplicate nickname"))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerError(responseInfo models.MessageInfo, paramsValues error) error {
	command := constants.CGCommands.ResponseServerError

	params, err := models.CreateParams(command.ParamsNames, []string{paramsValues.Error()})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	connection := responseInfo.ConnectionInfo.Connection
	playerName := responseInfo.PlayerNickname

	//convert to message
	message := models.CreateMessage(playerName, command.CommandID, params)

	paramsValues = connectionWrite(connection, message)
	if paramsValues != nil {
		return fmt.Errorf("error writing %w", paramsValues)
	}

	//disconnect player
	err = CloseConnection(connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}

	return nil
}

func SendResponseServerGameList(responseInfo models.MessageInfo, paramsValues []*models.Game) error {
	command := constants.CGCommands.ResponseServerGameList

	value := parser.ConvertListGameListToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error creating params %w", err)
	}
	connection := responseInfo.ConnectionInfo.Connection
	playerName := responseInfo.PlayerNickname

	//convert to message
	message := models.CreateMessage(playerName, command.CommandID, params)

	err = connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing %w", err)
	}

	return nil
}

func SendResponseServerSelectCubes(values []int, player *models.Player, message models.Message) error {
	command := constants.CGCommands.ResponseServerSelectCubes

	err := sendResponseServerSelectCubes(command, values, player, message)
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

	err := sendMessage(responseInfo, command, constants.CGNetworkEmptyParams)
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

	err := sendMessage(responseInfo, command, constants.CGNetworkEmptyParams)
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
	time.Sleep(1 * time.Second)

	//todo remove
	//reverse the order of the list
	newList := make([]*models.Player, len(playerList))
	for i := 0; i < len(playerList); i++ {
		newList[i] = playerList[len(playerList)-1-i]
	}
	playerList = newList

	for _, player := range playerList {
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}
	return nil
}

// CommunicationServerUpdateGameList
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

func CommunicationServerUpdatePlayerList(playerList []*models.Player) error {
	command := constants.CGCommands.ServerUpdatePlayerList

	paramsValue := parser.ConvertListPlayerListToNetworkString(playerList)
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

func CommunicationServerUpdateGameData(gameData models.GameData, playerList []*models.Player) error {
	command := constants.CGCommands.ServerUpdateGameData
	paramsValue := parser.ConvertListGameDataToNetworkString(gameData)
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
func CommunicationServerStartTurn(player *models.Player) (bool, error) {
	command := constants.CGCommands.ServerStartTurn

	isSuccess, err := sendMessageWithSuccessResponse(player, command, constants.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return isSuccess, fmt.Errorf("error sending ping player %w", err)
	}
	return isSuccess, err
}

func CommunicationServerPingPlayer(player *models.Player, expectedCommand constants.Command) (models.Message, bool, error) {
	command := constants.CGCommands.ServerPingPlayer
	params := constants.CGNetworkEmptyParams

	responseInfo := models.MessageInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendMessage(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, false, fmt.Errorf("error sending response %w", err)
	}

	connection := responseInfo.ConnectionInfo.Connection

	//wait for client response

	clientResponse, isTimeout, err := ConnectionReadTimeout(connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, false, fmt.Errorf("error reading %w", err)
	}
	if isTimeout {
		err = HandleTimeout(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error handling timeout %w", err)
		}
		return models.Message{}, false, nil
	}

	if clientResponse.CommandID == expectedCommand.CommandID {
		return clientResponse, true, nil
	}

	//check if client responded with success
	if !IsClientResponseCommand(clientResponse, responseInfo.PlayerNickname, constants.CGCommands.ResponseClientSuccess) {
		return models.Message{}, false, HandleClienResponseNotSuccess(player)
	}

	return models.Message{}, true, nil

	//isSuccess, err := sendMessageWithSuccessResponse(player, command, constants.CGNetworkEmptyParams)
	//if err != nil {
	//	errorHandeling.PrintError(err)
	//	return isSuccess, fmt.Errorf("error sending ping player %w", err)
	//}
	//return isSuccess, err
}

// CommunicationServerReconnectGameList
func CommunicationServerReconnectGameList(player *models.Player, paramsValues []*models.Game) (bool, error) {
	command := constants.CGCommands.ServerReconnectGameList

	value := parser.ConvertListGameListToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error creating params %w", err)
	}

	isSuccess, err := sendMessageWithSuccessResponse(player, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error sending ping player %w", err)
	}

	return isSuccess, err
}

//region RECONNECT MESSAGES

// CommunicationServerReconnectGameData
func CommunicationServerReconnectGameData(player *models.Player, paramsValues models.GameData) (bool, error) {
	command := constants.CGCommands.ServerReconnectGameData

	value := parser.ConvertListGameDataToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error creating params %w", err)
	}

	isSuccess, err := sendMessageWithSuccessResponse(player, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error sending ping player %w", err)
	}

	return isSuccess, err
}

// CommunicationServerReconnectPlayerList
func CommunicationServerReconnectPlayerList(player *models.Player, paramsValues []*models.Player) (bool, error) {
	command := constants.CGCommands.ServerReconnectPlayerList

	value := parser.ConvertListPlayerListToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error creating params %w", err)
	}

	isSuccess, err := sendMessageWithSuccessResponse(player, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error sending ping player %w", err)
	}

	return isSuccess, err
}

//endregion
//endregion

//endregion

//region RECEIVE FUNCTIONS

func CommunicationReadContinouslyWithPing(player *models.Player, command constants.Command) (models.Message, bool, error) {

	connection := player.GetConnectionInfo().Connection

	clientResponse, err := readContinouslyWithPing(player, command)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, false, fmt.Errorf("error reading %w", err)
	}
	if clientResponse.CommandID == constants.CGCommands.ResponseClientSuccess.CommandID {
		return clientResponse, true, nil
	}
	//check if client send ClientRollDice
	if clientResponse.CommandID != command.CommandID {
		err := helpers.RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error removing player from lists %w", err)
		}

		err = CloseConnection(connection)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error closing connection %w", err)
		}
		return models.Message{}, false, nil
	}

	return clientResponse, true, nil
}

func CommunicationReadClient(player *models.Player) (models.Message, bool, error) {

	connection := player.GetConnectionInfo().Connection

	clientResponse, err := communicationRead(player, connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, false, fmt.Errorf("error reading %w", err)
	}

	clientEndTurn := constants.CGCommands.ClientEndTurn.CommandID
	clientSelectedCubes := constants.CGCommands.ClientSelectedCubes.CommandID

	//check if client send ClientRollDice
	if clientResponse.CommandID != clientEndTurn && clientResponse.CommandID != clientSelectedCubes {
		err := helpers.RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error removing player from lists %w", err)
		}

		err = CloseConnection(connection)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error closing connection %w", err)
		}
		return models.Message{}, false, nil
	}

	return clientResponse, true, nil
}

//endregion

//endregion

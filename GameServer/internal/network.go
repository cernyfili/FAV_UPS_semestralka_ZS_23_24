package internal

import (
	"bufio"
	"errors"
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/utils"
	"gameserver/internal/utils/errorHandeling"
	"net"
	"os"
	"time"
)

// region GLOBAL VARIABLES
const (
	connType = "tcp"
	connHost = "127.0.0.1"
	connPort = "10000"
)
const cTimeout = time.Second

var gReceivedMessageList = models.CreateMessageList(models.Received)
var gSendMessageList = models.CreateMessageList(models.Send)

//endregion

// enum
type ResponseResult int

const (
	Succes ResponseResult = iota
	Timeout
	WrongResponse
	Error
)

// StartServer starts a TCP server that listens on the specified port.
func StartServer() {
	ln, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		errorHandeling.PrintError(err)
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server is listening on port " + connPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			errorHandeling.PrintError(err)
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// region PRIVATE FUNCTIONS
func sendResponseServerDice(command utils.Command, values []int, player *models.Player, message models.Message) error {
	paramsValue := ConvertCubeValuesToNetworkString(values)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: models.ConnectionInfo{
			Connection: player.GetConnectionInfo().Connection,
			TimeStamp:  message.TimeStamp,
		},
		PlayerNickname: message.PlayerNickname,
	}

	err = sendResponse(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func sendResponse(responseInfo models.NetworkResponseInfo, command utils.Command, params []utils.Params) error { //todo refactor where used
	connection := responseInfo.ConnectionInfo.Connection

	//convert to message
	message := models.CreateResponseMessage(responseInfo, command.CommandID, params)

	err := connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing %w", err)
	}
	return nil
}

// sendMessageWithSuccessResponse
// Return: bool isSuccess - if communication was successful
func sendMessageWithSuccessResponse(player *models.Player, command utils.Command, params []utils.Params) (bool, error) {
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendResponse(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error sending response %w", err)
	}

	connection := responseInfo.ConnectionInfo.Connection

	//wait for client response
	timeout := cTimeout
	clientResponse, isTimeout, err := connectionReadTimeout(connection, timeout)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error reading %w", err)
	}
	if isTimeout {
		err = handleTimeout(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("error handling timeout %w", err)
		}
		return false, nil
	}

	//check if client responded with success
	if !isClientResponseCommand(clientResponse, responseInfo, utils.CGCommands.ResponseClientSuccess) {
		err := RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("error removing player from lists %w", err)
		}

		err = connection.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			return false, fmt.Errorf("error closing connection %w", err)
		}

		return false, nil
	}

	return true, nil
}

func sendUpdateList(player *models.Player, command utils.Command, params []utils.Params) error {

	canFire, err := player.GetStateMachine().CanFire(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error checking if can fire %w", err)
	}
	if !canFire {
		return nil
	}

	connection := player.GetConnectionInfo().Connection
	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err = sendResponse(responseInfo, command, params)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	//wait for client response
	timeout := cTimeout
	clientResponse, isTimeout, err := connectionReadTimeout(connection, timeout)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error reading %w", err)
	}
	if isTimeout {
		err := handleTimeout(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error handling timeout %w", err)
		}
		return nil
	}

	//check if client responded with success
	if !isClientResponseCommand(clientResponse, responseInfo, utils.CGCommands.ResponseClientSuccess) {
		err := RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error removing player from lists %w", err)
		}

		err = connection.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error closing connection %w", err)
		}
		return nil
	}

	err = player.FireStateMachine(command.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error firing state machine %w", err)
	}

	return nil
}

func communicationRead(player *models.Player, connection net.Conn) (models.Message, models.Message, error) {
	timeout := cTimeout

	var clientResponse models.Message
	var isTimeout bool
	var err error
	doesRespond := true
	for doesRespond {
		//reading clientRollDice
		clientResponse, isTimeout, err = connectionReadTimeout(connection, timeout)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, models.Message{}, fmt.Errorf("error reading %w", err)
		}
		if isTimeout {
			isSuccess, err := CommunicationServerPingPlayer(player)
			if err != nil {
				errorHandeling.PrintError(err)
				return models.Message{}, models.Message{}, fmt.Errorf("error sending ping player %w", err)
			}
			if !isSuccess {
				doesRespond = false
				break
			}
		}
	}
	return clientResponse, models.Message{}, nil
}

func handlePlayerTimeOut(connection net.Conn) error {
	err := connection.Close()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}
	return nil
}

func handleConnection(conn net.Conn) {

	//todo while loop for reading new messages
	//todo handle timeout and ping
	logger.Log.Info("New connection from " + conn.RemoteAddr().String())
	message, err := connectionRead(conn)
	if err != nil {
		errorHandeling.PrintError(err)

		fmt.Println("Error reading:", err)
		//errorHandeling.PrintError("Error reading:", err)
		err := conn.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			//errorHandeling.PrintError("Error closing:", err)
			fmt.Println("Error closing:", err)
			return
		}
		return
	}

	/*var response string*/

	err = ProcessMessage(message, conn)
	if err != nil {
		errorHandeling.PrintError(err)
		fmt.Println("Error processing message:", err)
		return
	}
}

func handleTimeout(player *models.Player) error {
	commandError := utils.CGCommands.ErrorPlayerUnreachable
	err := player.FireStateMachine(commandError.Trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error firing state machine %w", err)
	}

	player.SetConnected(false)

	return nil
}

func connectionReadTimeout(connection net.Conn, timeout time.Duration) (models.Message, bool, error) {
	isTimeout := false

	// Set the timeout
	deadline := time.Now().Add(timeout)
	err := connection.SetReadDeadline(deadline)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, fmt.Errorf("error setting read deadline: %w", err)
	}

	buffer := make([]byte, 1024)
	_, err = connection.Read(buffer)
	if err != nil {
		errorHandeling.PrintError(err)
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			isTimeout = true
		}
		return models.Message{}, isTimeout, nil
	}
	messageStr := string(buffer)

	message, err := ParseMessage(messageStr)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, fmt.Errorf("Error parsing message: %w", err)
	}

	//Save to logger
	err = gReceivedMessageList.AddItem(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, isTimeout, err
	}

	return message, isTimeout, nil
}

func connectionRead(connection net.Conn) (models.Message, error) {

	reader := bufio.NewReader(connection)
	var messageStr string

	for {

		line, err := reader.ReadString(utils.CMessageEndDelimiter)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, fmt.Errorf("error reading: %w", err)
		}
		messageStr += line

		// Check if the message ends with a newline character
		if len(line) > 0 && line[len(line)-1] == '\n' {
			break
		}
		if len(messageStr) > utils.CMaxMessageSize {
			return models.Message{}, fmt.Errorf("message is too long")
		}
	}

	//logger.Log.Info("Received message: " + messageStr)

	message, err := ParseMessage(messageStr)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, fmt.Errorf("Error parsing message:", err)
	}

	err = gReceivedMessageList.AddItem(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, err
	}

	return message, nil
}

func connectionWrite(connection net.Conn, message models.Message) error {
	messageStr, err := ConvertMessageToNetworkString(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	_, err = connection.Write([]byte(messageStr))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing", err)
	}

	err = gSendMessageList.AddItem(message)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func isClientResponseCommand(clientResponse models.Message, originalInfo models.NetworkResponseInfo, command utils.Command) bool {
	if clientResponse.PlayerNickname != originalInfo.PlayerNickname {
		return false
	}

	if clientResponse.TimeStamp != originalInfo.ConnectionInfo.TimeStamp {
		return false
	}

	if clientResponse.CommandID != command.CommandID {
		return false
	}

	return true
}

//endregion

//region SEND RESPONSE FUNCTIONS

func SendResponseServerSuccess(responseInfo models.NetworkResponseInfo) error {
	//convert to message
	message := models.CreateResponseMessage(responseInfo, utils.CGCommands.ResponseServerSuccess.CommandID, utils.CGNetworkEmptyParams)

	err := connectionWrite(responseInfo.ConnectionInfo.Connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error writing %w", err)
	}

	return nil
}

func SendResponseServerErrDuplicitNickname(responseInfo models.NetworkResponseInfo) error {
	err := SendResponseServerError(responseInfo, fmt.Errorf("error duplicate nickname"))
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerError(responseInfo models.NetworkResponseInfo, paramsValues error) error {
	command := utils.CGCommands.ResponseServerError

	params, err := models.CreateParams(command.ParamsNames, []string{paramsValues.Error()})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	connection := responseInfo.ConnectionInfo.Connection

	//convert to message
	message := models.CreateResponseMessage(responseInfo, command.CommandID, params)

	paramsValues = connectionWrite(connection, message)
	if paramsValues != nil {
		return fmt.Errorf("error writing %w", paramsValues)
	}

	//disconnect player
	err = connection.Close()
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error closing connection %w", err)
	}

	return nil
}

func SendResponseServerGameList(responseInfo models.NetworkResponseInfo, paramsValues []*models.Game) (bool, error) {
	command := utils.CGCommands.ResponseServerGameList

	value := ConvertGameListToNetworkString(paramsValues)

	params, err := models.CreateParams(command.ParamsNames, []string{value})
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error creating params %w", err)
	}
	connection := responseInfo.ConnectionInfo.Connection

	//convert to message
	message := models.CreateResponseMessage(responseInfo, command.CommandID, params)

	err = connectionWrite(connection, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error writing %w", err)
	}

	return true, nil
}

func SendResponseServerDiceNext(values []int, player *models.Player, message models.Message) error {
	command := utils.CGCommands.ResponseServerDiceNext

	err := sendResponseServerDice(command, values, player, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerDiceEndTurn(values []int, player *models.Player, message models.Message) error {
	command := utils.CGCommands.ResponseServerDiceEndTurn

	err := sendResponseServerDice(command, values, player, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}
	return nil
}

func SendResponseServerNextDiceSuccess(player *models.Player) error {
	command := utils.CGCommands.ResponseServerNextDiceSuccess

	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendResponse(responseInfo, command, utils.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

func SendResponseServerNextDiceEndScore(player *models.Player) error {
	command := utils.CGCommands.ResponseServerNextDiceEndScore

	responseInfo := models.NetworkResponseInfo{
		ConnectionInfo: player.GetConnectionInfo(),
		PlayerNickname: player.GetNickname(),
	}

	err := sendResponse(responseInfo, command, utils.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("error sending response %w", err)
	}

	return nil
}

//endregion

//region COMMUNICATION FUNCTIONS

// CommunicationServerReconnectGameList
func CommunicationServerReconnectGameList(player *models.Player, paramsValues []*models.Game) (bool, error) {
	command := utils.CGCommands.ServerReconnectGameList

	value := ConvertGameListToNetworkString(paramsValues)

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

// CommunicationServerReconnectGameData
func CommunicationServerReconnectGameData(player *models.Player, paramsValues models.GameData) (bool, error) {
	command := utils.CGCommands.ServerReconnectGameData

	value := ConvertGameDataToNetworkString(paramsValues)

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
	command := utils.CGCommands.ServerReconnectPlayerList

	value := ConvertPlayerListToNetworkString(paramsValues)

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

func CommunicationServerUpdateGameList(playerList []*models.Player, paramsValues []*models.Game) error {
	command := utils.CGCommands.ServerUpdateGameList
	paramsValue := ConvertGameListToNetworkString(paramsValues)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	for _, player := range playerList {
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}

	return nil

}

func CommunicationServerUpdatePlayerList(playerList []*models.Player) error {
	command := utils.CGCommands.ServerUpdatePlayerList

	paramsValue := ConvertPlayerListToNetworkString(playerList)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	for _, player := range playerList {
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}

	return nil
}

func CommunicationServerUpdateStartGame(playerList []*models.Player) error {
	command := utils.CGCommands.ServerUpdateStartGame
	params := utils.CGNetworkEmptyParams

	for _, player := range playerList {
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}

	return nil
}

func CommunicationServerUpdateGameData(gameData models.GameData, playerList []*models.Player) error {
	command := utils.CGCommands.ServerUpdateGameData
	paramsValue := ConvertGameDataToNetworkString(gameData)
	params, err := models.CreateParams(command.ParamsNames, []string{paramsValue})
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	for _, player := range playerList {
		err := sendUpdateList(player, command, params)
		if err != nil {
			errorHandeling.PrintError(err)
			return fmt.Errorf("error sending update list %w", err)
		}
	}

	return nil
}

// CommunicationServerUpdateGameData
func CommunicationServerStartTurn(player *models.Player) (bool, error) {
	command := utils.CGCommands.ServerStartTurn

	isSuccess, err := sendMessageWithSuccessResponse(player, command, utils.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return isSuccess, fmt.Errorf("error sending ping player %w", err)
	}
	return isSuccess, err
}

func CommunicationServerPingPlayer(player *models.Player) (bool, error) {
	command := utils.CGCommands.ServerPingPlayer

	isSuccess, err := sendMessageWithSuccessResponse(player, command, utils.CGNetworkEmptyParams)
	if err != nil {
		errorHandeling.PrintError(err)
		return isSuccess, fmt.Errorf("error sending ping player %w", err)
	}
	return isSuccess, err
}

func CommunicationReadClientRollDice(player *models.Player) (models.Message, bool, error) {
	command := utils.CGCommands.ClientRollDice

	connection := player.GetConnectionInfo().Connection

	clientResponse, message, err := communicationRead(player, connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return message, false, fmt.Errorf("error reading %w", err)
	}

	//check if client send ClientRollDice
	if clientResponse.CommandID != command.CommandID {
		err := RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error removing player from lists %w", err)
		}

		err = connection.Close()
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

	clientResponse, message, err := communicationRead(player, connection)
	if err != nil {
		errorHandeling.PrintError(err)
		return message, false, fmt.Errorf("error reading %w", err)
	}

	clientEndTurn := utils.CGCommands.ClientEndTurn.CommandID
	clientNextDice := utils.CGCommands.ClientNextDice.CommandID

	//check if client send ClientRollDice
	if clientResponse.CommandID != clientEndTurn && clientResponse.CommandID != clientNextDice {
		err := RemovePlayerFromGame(player)
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error removing player from lists %w", err)
		}

		err = connection.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			return models.Message{}, false, fmt.Errorf("error closing connection %w", err)
		}
		return models.Message{}, false, nil
	}

	return clientResponse, true, nil
}

//endregion

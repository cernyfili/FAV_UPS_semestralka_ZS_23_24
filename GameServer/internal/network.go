package internal

import (
	"fmt"
	"gameserver/internal/utils"
	"net"
	"os"
	"time"
)

// region GLOBAL VARIABLES
const (
	connType = "tcp"
	connHost = "0.0.0.0"
	connPort = "10000"
)
const cTimeout = time.Second

var gMessageList = utils.GetInstanceMessageList()
var gResponseList = utils.GetInstanceMessageList()

//endregion

//todo handle network error client doesnt get message - array of history of responses

// StartServer starts a TCP server that listens on the specified port.
func StartServer(port string) {
	ln, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server is listening on port " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// region PRIVATE FUNCTIONS
func handleConnection(conn net.Conn) {

	message, err := connectionRead(conn)
	if err != nil {
		fmt.Println("Error reading:", err)
		conn.Close()
		return
	}

	/*var response string*/

	err = ProcessMessage(message, conn)
	if err != nil {
		/*response = err.Error()*/
		return
	}
	/*
		response = cSuccesMessage

		// Respond to the client
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing:", err)
		}

		conn.Close()*/
}

func connectionReadTimeout(connection net.Conn, timeout time.Duration) (utils.Message, error) {
	// Set the timeout
	deadline := time.Now().Add(timeout)
	err := connection.SetReadDeadline(deadline)
	if err != nil {
		return utils.Message{}, fmt.Errorf("error setting read deadline: %w", err)
	}

	buffer := make([]byte, 1024)
	_, err = connection.Read(buffer)
	if err != nil {
		return utils.Message{}, fmt.Errorf("error reading", err)
	}
	messageStr := string(buffer)

	message, err := ParseMessage(messageStr)
	if err != nil {
		return utils.Message{}, fmt.Errorf("Error parsing message:", err)
	}

	//Save to logger
	err = gMessageList.AddItem(message)
	if err != nil {
		return utils.Message{}, err
	}

	return message, nil
}

func connectionRead(connection net.Conn) (utils.Message, error) {

	buffer := make([]byte, 1024)
	_, err := connection.Read(buffer)
	if err != nil {
		return utils.Message{}, fmt.Errorf("error reading", err)
	}
	messageStr := string(buffer)

	message, err := ParseMessage(messageStr)
	if err != nil {
		return utils.Message{}, fmt.Errorf("Error parsing message:", err)
	}

	//Save to logger
	err = gMessageList.AddItem(message)
	if err != nil {
		return utils.Message{}, err
	}

	return message, nil
}

func connectionWrite(connection net.Conn, message utils.Message) error {
	messageStr, err := ConvertMessageToNetworkString(message)
	if err != nil {
		return err
	}

	_, err = connection.Write([]byte(messageStr))
	if err != nil {
		return fmt.Errorf("error writing", err)
	}

	err = gResponseList.AddItem(message)
	if err != nil {
		return err
	}

	return nil
}

func isClientResponseSuccess(clientResponse utils.Message, originalInfo utils.NetworkResponseInfo) bool {
	if clientResponse.PlayerNickname != originalInfo.PlayerNickname {
		return false
	}

	if clientResponse.TimeStamp != originalInfo.ConnectionInfo.TimeStamp {
		return false
	}

	if clientResponse.CommandID != cCommands.ResponseServerSuccess.CommandID {
		return false
	}

	return true
}

//endregion

//region SEND RESPONSE FUNCTIONS

func SendResponseSuccess(responseInfo utils.NetworkResponseInfo) error {
	//convert to message
	message := utils.CreateResponseMessage(responseInfo, cCommands.ResponseServerSuccess.CommandID, utils.CNetoworkEmptyParams)

	err := connectionWrite(responseInfo.ConnectionInfo.Connection, message)
	if err != nil {
		return fmt.Errorf("error writing %s", err)
	}

	return nil
}

func SendResponseDuplicitNickname(responseInfo utils.NetworkResponseInfo) error {
	//convert to message
	message := utils.CreateResponseMessage(responseInfo, cCommands.ResponseServerErrDuplicitNickname.CommandID, utils.CNetoworkEmptyParams)

	err := connectionWrite(responseInfo.ConnectionInfo.Connection, message)
	if err != nil {
		return fmt.Errorf("error writing %s", err)
	}

	//wait for client response
	connection := responseInfo.ConnectionInfo.Connection
	timeout := cTimeout
	clientResponse, err := connectionReadTimeout(connection, timeout)
	if err != nil {
		//if client doesnt respond
		err := connection.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("error reading %s", err)
	}

	//check if client responded with success
	if !isClientResponseSuccess(clientResponse, responseInfo) {
		err := connection.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("client didnt respond with same nickname")
	}

	return nil
}

//endregion

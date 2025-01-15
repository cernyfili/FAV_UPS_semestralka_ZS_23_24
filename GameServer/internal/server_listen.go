package internal

import (
	"fmt"
	"gameserver/internal/command_processing"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/network"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"net"
	"os"
)

func StartServer() {
	RunServer()
}

// RunServer starts a TCP server that listens on the specified port.
func RunServer() {
	ln, err := net.Listen(constants.CConnType, constants.CConnHost+":"+constants.CConnPort)
	if err != nil {
		errorHandeling.PrintError(err)
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer func() {
		//todo remove
		errorHandeling.PrintError(fmt.Errorf("Error closing listener:", ln.Close()))
		err := ln.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			fmt.Println("Error closing listener:", err)
		}
	}()
	fmt.Println("Server is listening on port " + constants.CConnPort)

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

func getMessagePlayer(message models.Message) (*models.Player, error) {
	playerList := models.GetInstancePlayerList()
	player, err := playerList.GetItem(message.PlayerNickname)
	if err != nil {
		return nil, fmt.Errorf("Error getting player: %w", err)
	}
	return player, nil
}

func handleConnection(conn net.Conn) {
	logger.Log.Info("New connection from " + conn.RemoteAddr().String())
	//todo handle timeout and ping
	for {
		// if connection is closed
		if conn == nil {
			return
		}

		//message, err := connectionRead(conn)
		message, isTimeout, err := network.ConnectionReadTimeout(conn)

		var isClientResponseSuccessExpected bool
		if err != nil {
			errorHandeling.PrintError(err)
			//todo remove

			fmt.Println("Error reading:", err)
			//errorHandeling.PrintError("Error reading:", err)
			err := network.CloseConnection(conn)
			if err != nil {
				errorHandeling.PrintError(err)
				//errorHandeling.PrintError("Error closing:", err)
				fmt.Println("Error closing:", err)
				return
			}
			return
		}

		var playerMessage *models.Player
		playerMessage, err = getMessagePlayer(message)
		if err != nil {
			isClientResponseSuccessExpected = false
		} else {
			isClientResponseSuccessExpected = playerMessage.IsResponseSuccessExpected()
		}

		if isTimeout {
			if isClientResponseSuccessExpected {
				err = network.HandleTimeout(playerMessage)
				if err != nil {
					errorHandeling.PrintError(err)
					fmt.Println("Error handling timeout:", err)
					return
				}
				err = fmt.Errorf("Timeout for player expected ClientResponseSuccess")
				errorHandeling.PrintError(err)
				return
			}

			// StartTurn
			player := models.GetInstancePlayerList().GetPlayerByConnection(conn)
			if player != nil {
				playerGame := models.GetInstanceGameList().GetPlayersGame(player)
				if playerGame != nil {
					isPlayerTurn := playerGame.IsPlayerTurn(player)
					if isPlayerTurn {
						isNextPlayerTurn, err := command_processing.ProcessPlayerTurn(playerGame)
						if err != nil {
							errorHandeling.PrintError(err)
							fmt.Println("Error processing player turn:", err)
							return
						}
						if !isNextPlayerTurn {
							err := fmt.Errorf("Error processing player turn: isNextPlayerTurn is false")
							errorHandeling.PrintError(err)
							fmt.Println("Error processing player turn: isNextPlayerTurn is false")
							return
						}
						err = playerGame.ShiftTurn()
						if err != nil {
							errorHandeling.PrintError(err)
							fmt.Println("Error shifting turn:", err)
							return
						}
					}
				}
			}

			continue
		}

		if isClientResponseSuccessExpected {
			if message.CommandID != constants.CGCommands.ResponseClientSuccess.CommandID {
				err = network.HandleClienResponseNotSuccess(playerMessage)
				if err != nil {
					errorHandeling.PrintError(err)
					fmt.Println("Error handling client response not success:", err)
					return
				}
				err = fmt.Errorf("Timeout for player expected ClientResponseSuccess")
				errorHandeling.PrintError(err)
				return
			}

			playerMessage.DecreaseResponseSuccessExpected()
			continue
		}

		err = command_processing.ProcessMessage(message, conn)
		if err != nil {
			errorHandeling.PrintError(err)
			fmt.Println("Error processing message:", err)
			return
		}

	}
}

//func connectionRead(connection net.Conn) (models.Message, error) {
//
//	reader := bufio.NewReader(connection)
//	var messageStr string
//
//	for {
//
//		line, err := reader.ReadString(constants.CMessageEndDelimiter)
//		if err != nil {
//			errorHandeling.PrintError(err)
//			return models.Message{}, fmt.Errorf("error reading: %w", err)
//		}
//		messageStr += line
//
//		// Check if the message ends with a newline character
//		if len(line) > 0 && line[len(line)-1] == '\n' {
//			break
//		}
//		if len(messageStr) > constants.CMessageMaxSize {
//			return models.Message{}, fmt.Errorf("message is too long")
//		}
//	}
//
//	//logger.Log.Info("Received message: " + messageStr)
//
//	message, err := parser.ParseMessage(messageStr)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return models.Message{}, fmt.Errorf("Error parsing message:", err)
//	}
//
//	err = network_utils.GReceivedMessageList.AddItem(message)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return models.Message{}, err
//	}
//
//	return message, nil
//}

package internal

import (
	"fmt"
	"gameserver/internal/command_processing"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/models/state_machine"
	"gameserver/internal/network"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"net"
	"os"
	"time"
)

func StartServer() {
	RunServer()
}

// RunServer starts a TCP server that listens on the specified port.
func RunServer() {
	ln, err := net.Listen(constants.CConnType, constants.CConIPadress+":"+constants.CConnPort)
	if err != nil {
		errorHandeling.PrintError(err)
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer func() {
		errorHandeling.PrintError(fmt.Errorf("Error closing listener:", ln.Close()))
		err := ln.Close()
		if err != nil {
			errorHandeling.PrintError(err)
			fmt.Println("Error closing listener:", err)
		}
	}()
	fmt.Println("Server is listening on " + constants.CConIPadress + ":" + constants.CConnPort)

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

//func getMessagePlayer(message models.Message) (*models.Player, error) {
//	playerList := models.GetInstancePlayerList()
//	player, err := playerList.GetItem(message.PlayerNickname)
//	if err != nil {
//		return nil, fmt.Errorf("Error getting player: %w", err)
//	}
//	return player, nil
// }

func _tryStartTurn(conn net.Conn) error {
	player := models.GetInstancePlayerList().GetPlayerByConnection(conn)
	if player == nil {
		return nil
	}

	game := models.GetInstanceGameList().GetPlayersGame(player)
	if game == nil {
		return nil
	}

	if !game.IsPlayerTurn(player) {
		return nil
	}

	if player.IsInTurn() {
		//already have been send serverstart
		return nil
	}

	//is player turn and havent been started yet
	err := command_processing.ProcessPlayerTurn(game)
	if err != nil {
		err = fmt.Errorf("Error processing start turn: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

func handleConnection(conn net.Conn) {
	logger.Log.Info("New connection from " + conn.RemoteAddr().String())
	go continuousSendPing(conn)

	for {
		// if connection is closed
		if conn == nil {
			return
		}
		player := models.GetInstancePlayerList().GetPlayerByConnection(conn)

		//isConnected := checkTotalDisconnect(player)
		//logger.Log.Debugf("TOTAL_DISCONNECT: Check Total disconect IsConnected: " + fmt.Sprint(isConnected))
		//if !isConnected {
		//	return
		//}

		// client havent responded to some of the messages
		isResponseTimeout, err := checkTimeoutResponseSuccess(player)
		if err != nil {
			err = fmt.Errorf("Error checking timeout: %w", err)
			errorHandeling.PrintError(err)
			return
		}
		if isResponseTimeout {
			err := disconnectPlayer(player)
			if err != nil {
				err = fmt.Errorf("Error disconnecting player: %w", err)
				errorHandeling.PrintError(err)
				return
			}
			return
		}

		// StartTurn
		err = _tryStartTurn(conn)
		if err != nil {
			errorHandeling.PrintError(err)
			fmt.Println("Error starting turn:", err)
			return
		}

		messageList, isTimeout, err := network.Read(conn)
		if err != nil {
			err = fmt.Errorf("Error reading: %w", err)
			errorHandeling.PrintError(err)
			//if error when reading client login messsage
			if player == nil {
				err := network.CloseConnection(conn)
				if err != nil {
					err = fmt.Errorf("Error closing: %w", err)
					errorHandeling.PrintError(err)
					return
				}
			}

			//if error when reading client message
			err = disconnectPlayer(player)
			if err != nil {
				err = fmt.Errorf("Error disconnecting player: %w", err)
				errorHandeling.PrintError(err)
				return
			}
			//network.ImidiateDisconnectPlayer(player.GetNickname())

			return
		}
		if isTimeout {
			//if client havenot logeed yet - waiting for login
			if player == nil {
				continue
			}

			//if in reconnect mode
			if !player.IsConnected() {
				continue
			}

			if player.GetCurrentStateName() == state_machine.StateNameMap.StateReconnect {
				continue
			}

			//if client have logged in - reading with ping
			err := command_processing.ProcessSendPingPlayer(player)
			if err != nil {
				err = fmt.Errorf("Timeout for player: %w", err)
				errorHandeling.PrintError(err)
				return
			}
			continue
		}

		for _, message := range messageList {
			err = command_processing.ProcessMessage(message, conn)
			if err != nil {
				errorHandeling.PrintError(err)
				fmt.Println("Error processing message:", err)
				return
			}
		}
	}
}

func continuousSendPing(conn net.Conn) {
	for {
		if conn == nil {
			return
		}
		player := models.GetInstancePlayerList().GetPlayerByConnection(conn)

		if player == nil {
			continue
		}

		can_fire, err := player.GetStateMachine().CanFire(constants.CGCommands.ServerPingPlayer.Trigger)
		if err != nil {
			err = fmt.Errorf("Error checking if can fire: %w", err)
			errorHandeling.PrintError(err)
			return
		}
		if !can_fire {
			continue
		}

		err = command_processing.ProcessSendPingPlayer(player)
		if err != nil {
			logger.Log.Info("Couldne sending ping: " + err.Error())

			err = disconnectPlayer(player)
			if err != nil {
				err = fmt.Errorf("Error disconnecting player: %w", err)
				errorHandeling.PrintError(err)
				return
			}
			return
		}

		time.Sleep(constants.CPingTime)
	}
}

//func sendPing(player *models.Player) error {
//	if player == nil {
//		return nil
//	}
//
//	if !player.IsTimeForNewPing() {
//		return nil
//	}
//
//	err := command_processing.ProcessSendPingPlayer(player)
//	if err != nil {
//		err = fmt.Errorf("Error sending ping: %w", err)
//		errorHandeling.PrintError(err)
//		return err
//	}
//
//	player.SetLastPingCurrentTime()
//
//	return nil
//}

func disconnectPlayer(player *models.Player) error {
	if player == nil {
		errorHandeling.AssertError(fmt.Errorf("Error disconnecting player: player is nil"))
	}

	err := network.DisconnectPlayerConnection(player)
	if err != nil {
		err = fmt.Errorf("Error disconnecting player: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
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

func checkTimeoutResponseSuccess(player *models.Player) (bool, error) {
	if player == nil {
		return false, nil
	}

	if !player.IsResponseSuccessExpected() {
		return false, nil
	}

	isTimeout, err := player.IsResponseExpectedTimeout()
	if err != nil {
		errorHandeling.PrintError(err)
		return false, fmt.Errorf("error checking timeout %w", err)
	}

	return isTimeout, nil
}

func checkTotalDisconnect(player *models.Player) bool {
	if player == nil {
		return true
	}

	if player.IsConnected() {
		return true
	}

	if !player.IsTotalDisconnectTimeout() {
		return true
	}

	network.ImidiateDisconnectPlayer(player.GetNickname())

	return false
}

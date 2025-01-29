package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/pkg/stateless"
	"sync"
	"time"
)

//region DATA STRUCTURES

// struct response succes element

// Player represents a Player with a unique ID and nickname.
type Player struct {
	nickname                string
	isConnected             bool
	connectionInfo          ConnectionInfo
	stateMachine            *stateless.StateMachine
	responseSuccessExpected []Message
	mutex                   sync.Mutex
}

//endregion

// CreatePlayer creates a new Player with a unique ID and nickname.
func CreatePlayer(nickname string, isConnected bool, connectionInfo ConnectionInfo) *Player {
	return &Player{
		nickname:                nickname,
		isConnected:             isConnected,
		connectionInfo:          connectionInfo,
		responseSuccessExpected: []Message{},
		stateMachine:            CreateStateMachine(),
	}
}

//region GETTERS

// GetStateMachine returns the state machine of the Player
func (p *Player) GetStateMachine() *stateless.StateMachine {
	p.lock()
	defer p.unlock()

	return p.stateMachine
}

// GetNickname returns the nickname of the Player
func (p *Player) GetNickname() string {
	logger.Log.Debugf("Getting nickname of player ")

	p.lock()
	defer p.unlock()

	logger.Log.Debugf("Got nickname of player %s", p.nickname)

	return p.nickname
}

// Get connection info
func (p *Player) GetConnectionInfo() ConnectionInfo {
	p.lock()
	defer p.unlock()

	return p.connectionInfo
}

// IsConnected returns the connection status of the Player
func (p *Player) IsConnected() bool {
	p.lock()
	defer p.unlock()

	return p.isConnected
}

//func (p *Player) GetResponseSuccessExpected() int {
//	p.lock()
//	defer p.unlock()
//
//	return p.responseSuccessExpected
//}

// is player expecting a response
func (p *Player) IsResponseSuccessExpected() bool {
	p.lock()
	defer p.unlock()

	lenList := len(p.responseSuccessExpected)

	if lenList >= 2 {
		logger.Log.Infof("RESPONSE_EXPECTED: Player %s has %d responses expected", p.nickname, lenList)
	}

	return lenList > 0
}

//endregion

//region SETTERS

// Reset ResponseSuccessExpected
func (p *Player) resetResponseSuccessExpected() {
	logger.Log.Infof("RESPONSE_EXPECTED: Reseting Player %s has %d responses expected", p.nickname, len(p.responseSuccessExpected))

	p.responseSuccessExpected = []Message{}
}

// SetConnectionInfo sets the connection info of the Player
func (p *Player) SetConnectionInfo(connectionInfo ConnectionInfo) {
	p.lock()
	defer p.unlock()

	p.connectionInfo = connectionInfo
}

// SetConnected sets the connection status of the Player
func (p *Player) SetConnected(isConnected bool) {
	p.lock()
	defer p.unlock()

	p.isConnected = isConnected
}

// is response expected timeout
func (p *Player) IsResponseExpectedTimeout() (bool, error) {
	p.lock()
	defer p.unlock()

	currentTime := time.Now()
	timeOut := constants.CTimeout

	for _, message := range p.responseSuccessExpected {
		//convert string timestemp to time
		messageTime, err := time.Parse(constants.CMessageTimeFormat, message.TimeStamp)
		if err != nil {
			err = fmt.Errorf("Error parsing message time: %w", err)
			errorHandeling.PrintError(err)
			return false, err
		}

		isTimeout := currentTime.Sub(messageTime) > timeOut
		if isTimeout {
			return true, nil
		}

	}

	return false, nil
}

// increase the number of expected responses
func (p *Player) IncreaseResponseSuccessExpected(message Message) {
	p.lock()
	defer p.unlock()
	lenList := len(p.responseSuccessExpected)

	if lenList+1 >= 2 {
		logger.Log.Infof("RESPONSE_EXPECTED: When Increased Player %s has %d responses expected", p.nickname, lenList+1)
	}

	//Add to list
	p.responseSuccessExpected = append(p.responseSuccessExpected, message)
}

// decrease the number of expected responses
func (p *Player) DecreaseResponseSuccessExpected() error {
	p.lock()
	defer p.unlock()
	len_list := len(p.responseSuccessExpected)
	if len_list >= 2 {
		//player has that many responses expected
		logger.Log.Infof("RESPONSE_EXPECTED: Player %s has %d responses expected", p.nickname, len_list)
	}

	if len_list-1 < 0 {
		err := fmt.Errorf("Player has already expected a response")
		errorHandeling.PrintError(err)
		return err
	}

	//Remove last from list
	p.responseSuccessExpected = p.responseSuccessExpected[:len_list-1]
	return nil
}

// Fires the state machine
func (p *Player) FireStateMachine(trigger stateless.Trigger) error {
	p.lock()
	defer p.unlock()

	beforeState := p.stateMachine.MustState()

	err := p.stateMachine.Fire(trigger)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	afterState := p.stateMachine.MustState()

	//reset the expected responses
	if beforeState != afterState {
		//p.resetResponseSuccessExpected()
	}

	logger.Log.Infof("AUTOMATA: Player %s changed state from %s : - %s - : %s", p.nickname, beforeState, trigger, afterState)
	return nil
}

// lock the player
func (p *Player) lock() {
	//logger.Log.Infof("Locking player")
	p.mutex.Lock()
}

// unlock the player
func (p *Player) unlock() {
	//	logger.Log.Infof("Unlocking player")
	p.mutex.Unlock()
}

func (p *Player) IsInTurn() bool {
	allowedStates := []stateless.State{stateMyTurn, stateForkMyTurn, stateForkNextDice, stateNextDice}
	currentState := p.GetStateMachine().MustState()
	for _, state := range allowedStates {
		if currentState == state {
			return true
		}
	}
	return false
}

//endregion

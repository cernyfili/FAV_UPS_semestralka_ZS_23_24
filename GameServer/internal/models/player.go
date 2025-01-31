package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models/state_machine"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/pkg/stateless"
	"sync"
	"time"
)

//region DATA STRUCTURES

// struct connection states

type ConnectionStateType int

type ConnectionStateStruct struct {
	Connected       ConnectionStateType
	Disconnected    ConnectionStateType
	TotalDisconnect ConnectionStateType
}

var ConnectionStates = ConnectionStateStruct{
	Connected:       0,
	Disconnected:    1,
	TotalDisconnect: 2,
}

// Player represents a Player with a unique ID and nickname.
type Player struct {
	nickname                string
	connectionState         ConnectionStateType
	connectionInfo          ConnectionInfo
	stateMachine            *stateless.StateMachine
	responseSuccessExpected []Message
	mutex                   sync.Mutex
	lastPingTime            time.Time
}

//endregion

// CreatePlayer creates a new Player with a unique ID and nickname.
func CreatePlayer(nickname string, connectionInfo ConnectionInfo) *Player {
	return &Player{
		nickname:                nickname,
		connectionState:         ConnectionStates.Connected,
		connectionInfo:          connectionInfo,
		responseSuccessExpected: []Message{},
		stateMachine:            state_machine.CreateStateMachine(),
	}
}

//region GETTERS

// GetStateMachine returns the state machine of the Player
func (p *Player) GetStateMachine() *stateless.StateMachine {
	p.lock()
	defer p.unlock()

	return p.stateMachine
}

// GetConnectionState returns the connection state of the Player
func (p *Player) GetConnectionState() ConnectionStateType {
	p.lock()
	defer p.unlock()

	return p.connectionState
}

// GetNickname returns the nickname of the Player
func (p *Player) GetNickname() string {
	p.lock()
	defer p.unlock()

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

	return p.connectionState == ConnectionStates.Connected
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

// SetConnectedByBool sets the connection status of the Player
func (p *Player) SetConnectedByBool(isConnected bool) {
	p.lock()
	defer p.unlock()

	switch isConnected {
	case true:
		p.connectionState = ConnectionStates.Connected
	case false:
		p.connectionState = ConnectionStates.Disconnected
	}
}

func diffWithCurrentTime(timeStr string) time.Duration {
	messageTime, err := time.Parse(constants.CMessageTimeFormat, timeStr)
	if err != nil {
		errorHandeling.AssertError(fmt.Errorf("Error parsing message time: %w", err))
	}

	currentTime := time.Now()
	formattedCurrentTimeStr := currentTime.Format(constants.CMessageTimeFormat)
	formattedCurrentTime, err := time.Parse(constants.CMessageTimeFormat, formattedCurrentTimeStr)
	if err != nil {
		errorHandeling.AssertError(fmt.Errorf("Error parsing current time: %w", err))
	}

	diff := formattedCurrentTime.Sub(messageTime)

	return diff
}

// is response expected timeout
func (p *Player) IsResponseExpectedTimeout() (bool, error) {
	p.lock()
	defer p.unlock()

	timeOut := constants.CTimeout

	for _, message := range p.responseSuccessExpected {

		diff := diffWithCurrentTime(message.TimeStamp)

		isTimeout := diff > timeOut
		if isTimeout {
			logger.Log.Infof("TIME_OUT: - Response from Player %s for message %s", p.nickname, message.CommandID)
			//time values in log

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

	//todo add multiple to list not just the last
	//p.responseSuccessExpected = append(p.responseSuccessExpected, message)

	//crate new list where is only one element
	p.responseSuccessExpected = []Message{message}
}

// decrease the number of expected responses
func (p *Player) DecreaseResponseSuccessExpected(timeStamp string) error {
	p.lock()
	defer p.unlock()
	len_list := len(p.responseSuccessExpected)
	if len_list >= 2 {
		//player has that many responses expected
		logger.Log.Infof("RESPONSE_EXPECTED: Player %s has %d responses expected", p.nickname, len_list)
	}

	//if len_list-1 < 0 {
	//	err := fmt.Errorf("Player has already expected a response")
	//	errorHandeling.PrintError(err)
	//	return err
	//}

	//Remove with same timestamp
	//todo
	//for i, message := range p.responseSuccessExpected {
	//	if message.TimeStamp == timeStamp {
	//		p.responseSuccessExpected = append(p.responseSuccessExpected[:i], p.responseSuccessExpected[i+1:]...)
	//		return nil
	//	}
	//}

	if len_list == 0 {
		return nil
	}

	//remove last element
	p.responseSuccessExpected = p.responseSuccessExpected[:len(p.responseSuccessExpected)-1]
	return nil

	//err := fmt.Errorf("No response with timestamp %s found", timeStamp)
	//errorHandeling.PrintError(err)
	//return err
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
	//if beforeState != afterState {
	//	//p.resetResponseSuccessExpected()
	//}

	logger.Log.Infof("AUTOMATA: Player %s changed state from %s : - %s - : %s", p.nickname, beforeState, trigger, afterState)
	return nil
}

// Get current state name
func (p *Player) GetCurrentStateName() string {
	p.lock()
	defer p.unlock()

	currentState := p.stateMachine.MustState()
	currentStateName, ok := currentState.(string)
	if !ok {
		errorHandeling.AssertError(fmt.Errorf("cannot assert state name"))
	}

	return currentStateName
}

// Reset State Machine
func (p *Player) ResetStateMachine() {
	p.lock()
	defer p.unlock()

	p.stateMachine = state_machine.CreateStateMachine()
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
	allowedStatesName := []string{
		state_machine.StateNameMap.StateMyTurn,
		state_machine.StateNameMap.StateForkMyTurn,
		state_machine.StateNameMap.StateNextDice,
		state_machine.StateNameMap.StateForkNextDice,
	}

	currentStateName := p.GetCurrentStateName()

	for _, state := range allowedStatesName {
		if currentStateName == state {
			return true
		}
	}
	return false
}

func (p *Player) SetConnected(state ConnectionStateType) {
	p.lock()
	defer p.unlock()

	p.connectionState = state
}

func (p *Player) SetLastPingCurrentTime() {
	p.lock()
	defer p.unlock()

	p.lastPingTime = time.Now()
}

func (p *Player) IsTimeForNewPing() bool {
	p.lock()
	defer p.unlock()

	diff := time.Since(p.lastPingTime)

	return diff > constants.CPingTime
}

//endregion

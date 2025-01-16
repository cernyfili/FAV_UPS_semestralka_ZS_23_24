package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/pkg/stateless"
	"sync"
)

//region DATA STRUCTURES

// Player represents a Player with a unique ID and nickname.
type Player struct {
	nickname                string
	isConnected             bool
	connectionInfo          ConnectionInfo
	stateMachine            *stateless.StateMachine
	responseSuccessExpected int
	mutex                   sync.Mutex
}

//endregion

// CreatePlayer creates a new Player with a unique ID and nickname.
func CreatePlayer(nickname string, isConnected bool, connectionInfo ConnectionInfo) *Player {
	return &Player{
		nickname:                nickname,
		isConnected:             isConnected,
		connectionInfo:          connectionInfo,
		responseSuccessExpected: 0,
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
	logger.Log.Infof("Getting nickname of player ")

	p.lock()
	defer p.unlock()

	logger.Log.Infof("Got nickname of player %s", p.nickname)

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

	return p.responseSuccessExpected > 0
}

//endregion

//region SETTERS

// Reset ResponseSuccessExpected
func (p *Player) ResetResponseSuccessExpected() {
	p.lock()
	defer p.unlock()

	p.responseSuccessExpected = 0
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

// increase the number of expected responses
func (p *Player) IncreaseResponseSuccessExpected() {
	p.lock()
	defer p.unlock()
	//if p.responseSuccessExpected+1 > 1 {
	//	err := fmt.Errorf("Player has already expected a response")
	//	errorHandeling.PrintError(err)
	//}

	p.responseSuccessExpected++
}

// decrease the number of expected responses
func (p *Player) DecreaseResponseSuccessExpected() {
	p.lock()
	defer p.unlock()
	if p.responseSuccessExpected-1 < 0 {
		err := fmt.Errorf("Player has already expected a response")
		errorHandeling.PrintError(err)
	}

	p.responseSuccessExpected--
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

	logger.Log.Infof("Player %s changed state from %s : - %s - : %s", p.nickname, beforeState, trigger, afterState)
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

//endregion

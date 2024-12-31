package models

import (
	"gameserver/internal/logger"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/pkg/stateless"
	"sync"
)

//region DATA STRUCTURES

// Player represents a Player with a unique ID and nickname.
type Player struct {
	nickname       string
	game           *Game
	isConnected    bool
	connectionInfo ConnectionInfo
	stateMachine   *stateless.StateMachine
	mutex          sync.Mutex
}

//endregion

// CreatePlayer creates a new Player with a unique ID and nickname.
func CreatePlayer(nickname string, game *Game, isConnected bool, connectionInfo ConnectionInfo) *Player {
	return &Player{
		nickname:       nickname,
		game:           game,
		isConnected:    isConnected,
		connectionInfo: connectionInfo,
		stateMachine:   CreateStateMachine(),
	}
}

//region GETTERS

// GetStateMachine returns the state machine of the Player
func (p *Player) GetStateMachine() *stateless.StateMachine {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.stateMachine
}

// GetNickname returns the nickname of the Player
func (p *Player) GetNickname() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.nickname
}

// Get connection info
func (p *Player) GetConnectionInfo() ConnectionInfo {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.connectionInfo
}

func (p *Player) GetGame() *Game {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.game
}

// IsConnected returns the connection status of the Player
func (p *Player) IsConnected() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.isConnected
}

//endregion

//region SETTERS

// SetConnectionInfo sets the connection info of the Player
func (p *Player) SetConnectionInfo(connectionInfo ConnectionInfo) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.connectionInfo = connectionInfo
}

func (p *Player) SetGame(game *Game) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.game = game
}

// SetConnected sets the connection status of the Player
func (p *Player) SetConnected(isConnected bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.isConnected = isConnected
}

// Fires the state machine
func (p *Player) FireStateMachine(state stateless.State) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	beforeState := p.stateMachine.MustState()

	err := p.stateMachine.Fire(state)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	logger.Log.Infof("Player %s changed state from %s to %s", p.GetNickname(), beforeState, state)
	return nil
}

//endregion

package utils

import (
	"gameserver/pkg/stateless"
	"sync"
)

// Player represents a Player with a unique ID and nickname.
type Player struct {
	nickname       string
	game           *Game
	isConnected    bool
	connectionInfo ConnectionInfo
	stateMachine   *stateless.StateMachine
	mutex          sync.Mutex
}

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

// SetConnectionInfo sets the connection info of the Player
func (p *Player) SetConnectionInfo(connectionInfo ConnectionInfo) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.connectionInfo = connectionInfo
}

// Fires the state machine
func (p *Player) FireStateMachine(state stateless.State) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.stateMachine.Fire(state)
	if err != nil {
		return err
	}
	return nil
}

func (p *Player) SetGame(game *Game) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.game = game
}

func (p *Player) GetGame() *Game {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.game
}

// SetConnected sets the connection status of the Player
func (p *Player) SetConnected(isConnected bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.isConnected = isConnected
}

// IsConnected returns the connection status of the Player
func (p *Player) IsConnected() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.isConnected
}

//SetCo

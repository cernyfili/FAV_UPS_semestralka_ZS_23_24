package data_structure

import (
	"errors"
)

// enum for game state
type GameState int

const (
	Running GameState = iota
	Created
	Ended
)

// Player Data Structure represents plyaer data for game
type PlayerGameData struct {
	Player *Player
	Score  int
}

// Game represents a game with a unique ID, a list of g_players, and turn data.
type Game struct {
	GameID             int
	Name               string
	PlayersGameDataArr []PlayerGameData
	MaxPlayers         int
	TurnNum            int
	GameStateValue     GameState
}

// CreateGame creates a new game with a unique ID and initializes player and turn slices.
func CreateGame(name string, maxPlayers int) (*Game, error) {
	//Check if the arguments are valid
	if name == "" || maxPlayers <= 1 {
		return nil, errors.New("invalid arguments")
	}

	return &Game{
		Name:               name,
		PlayersGameDataArr: make([]PlayerGameData, 0),
		MaxPlayers:         maxPlayers,
		TurnNum:            0,
		GameStateValue:     Created,
	}, nil
}

// AddPlayer adds a new player to the game.
func (g *Game) AddPlayer(pPlayer *Player) error {
	if len(g.PlayersGameDataArr) >= g.MaxPlayers {
		return errors.New("game is full")
	}

	playerGameData := PlayerGameData{
		Player: pPlayer,
		Score:  0,
	}

	g.PlayersGameDataArr = append(g.PlayersGameDataArr, playerGameData)
	pPlayer.Game = g

	return nil
}

// NextTurn returns the next player in turn.
func (g *Game) NextTurn() (*Player, error) {
	if len(g.PlayersGameDataArr) == 0 {
		return nil, errors.New("no players in the game")
	}
	if g.GameStateValue != Running {
		return nil, errors.New("game is not running")
	}

	currentPlayerIndex := g.TurnNum % len(g.PlayersGameDataArr)
	nextPlayerIndex := (currentPlayerIndex + 1) % len(g.PlayersGameDataArr)

	g.TurnNum++

	return g.PlayersGameDataArr[nextPlayerIndex].Player, nil
}

func (g *Game) HasPlayer(player *Player) bool {
	for _, p := range g.PlayersGameDataArr {
		if p.Player == player {
			return true
		}
	}
	return false
}

func (g *Game) RemovePlayer(player *Player) error {
	for i, p := range g.PlayersGameDataArr {
		if p.Player == player {
			player.Game = nil
			g.PlayersGameDataArr = append(g.PlayersGameDataArr[:i], g.PlayersGameDataArr[i+1:]...)
			break
		}
	}
}

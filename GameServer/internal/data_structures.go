package internal

import "fmt"

// Player represents a player in the game.
type Player struct {
	PlayerID int
	Nickname string
}

// Turn represents a player's turn with their ID and score.
type Turn struct {
	PlayerID int
	Score    int
}

// Game represents a game with a unique ID, a list of players, and turn data.
type Game struct {
	GameID  string
	Players []Player
	Turns   []Turn
}

// CreateNewGame creates a new game with a unique ID and initializes player and turn slices.
func CreateNewGame(gameID string, players []Player) *Game {
	// Initialize the turn array with zero scores for each player.
	turns := make([]Turn, len(players))
	for i, player := range players {
		turns[i] = Turn{PlayerID: player.PlayerID, Score: 0}
	}
	return &Game{
		GameID:  gameID,
		Players: players,
		Turns:   turns,
	}
}

// AddScore updates the score for a specific player's turn.
func (g *Game) AddScore(playerID, score int) {
	for i, turn := range g.Turns {
		if turn.PlayerID == playerID {
			g.Turns[i].Score += score
			return
		}
	}
}

// DisplayGameInfo prints information about the game.
func (g *Game) DisplayGameInfo() {
	fmt.Printf("Game ID: %s\n", g.GameID)
	fmt.Println("Players:")
	for _, player := range g.Players {
		fmt.Printf("  Player ID: %d, Nickname: %s\n", player.PlayerID, player.Nickname)
	}
	fmt.Println("Turn Scores:")
	for _, turn := range g.Turns {
		fmt.Printf("  Player ID: %d, Score: %d\n", turn.PlayerID, turn.Score)
	}
}

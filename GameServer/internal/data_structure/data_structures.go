package data_structure

// Turn represents a player's turn with their ID and score.
type Turn struct {
	PlayerID int
	Score    int
}

// Player represents a player with a unique ID and nickname.
type Player struct {
	Nickname string
	Game     *Game
}

/*
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
	fmt.Println("PlayersGameDataArr:")
	for _, player := range g.PlayersGameDataArr {
		fmt.Printf("  Player ID: %d, Nickname: %s\n", player.PlayerID, player.Nickname)
	}
	fmt.Println("Turn Scores:")
	for _, turn := range g.Turns {
		fmt.Printf("  Player ID: %d, Score: %d\n", turn.PlayerID, turn.Score)
	}
}*/

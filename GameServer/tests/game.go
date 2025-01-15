package main

import (
	"fmt"
	"gameserver/internal/models"
)

func main() {

	game, err := models.CreateGame("game1", 2)
	if err != nil {
		panic(err)
	}
	err = game.AddPlayer(models.CreatePlayer("player1", true, models.ConnectionInfo{}))
	if err != nil {
		return
	}
	err = game.AddPlayer(models.CreatePlayer("player1", true, models.ConnectionInfo{}))
	if err != nil {
		return
	}
	//startgeam
	err = game.StartGame()
	if err != nil {
		return
	}

	//get player and get throw
	// Get the turn player
	turnPlayer, err := game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err := game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)
}

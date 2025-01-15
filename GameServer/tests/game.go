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
	err = game.AddPlayer(models.CreatePlayer("player2", true, models.ConnectionInfo{}))
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

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5
	var selectedCubeValues []int
	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}

	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}

	//next player
	err = game.NextPlayerTurn()
	if err != nil {
		panic(err)
	}
	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}
	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}
	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}
	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}
	//get player and get throw
	// Get the turn player
	turnPlayer, err = game.GetTurnPlayer()
	if err != nil {
		panic(err)
	}

	// Get a new throw for the turn player
	cubeValues, err = game.NewThrow(turnPlayer)
	if err != nil {
		panic(err)
	}

	// Print the cube values
	fmt.Println("Cube values:", cubeValues)

	// select from cub values values where is 1 or 5
	// Get the cube values that are 1 or 5

	for _, value := range cubeValues {
		if value == 1 || value == 5 {
			selectedCubeValues = append(selectedCubeValues, value)
		}
	}

	// set score
	err = game.SetPlayerScore(turnPlayer, selectedCubeValues)
	if err != nil {
		panic(err)
	}

}

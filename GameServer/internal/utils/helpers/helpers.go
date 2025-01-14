package helpers

import (
	"fmt"
	"gameserver/internal/models"
	"gameserver/internal/utils/errorHandeling"
)

func RemovePlayerFromLists(player *models.Player) error {

	playerlist := models.GetInstancePlayerList()

	gamelist := models.GetInstanceGameList()
	//Remove player from playerlist
	err := playerlist.RemoveItem(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create playersGame %w", err)
	}

	//Remove player from playersGame

	gameFromList := gamelist.GetPlayersGame(player)

	if !gameFromList.HasPlayer(player) {
		return fmt.Errorf("cannot create playersGame %w", err)
	}

	err = gameFromList.RemovePlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create playersGame %w", err)
	}

	return nil
}

func RemovePlayerFromGame(player *models.Player) error {
	err := RemovePlayerFromLists(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func RemovePlayerFromList(list []*models.Player, player *models.Player) []*models.Player {
	var newList []*models.Player
	for _, p := range list {
		if p != player {
			newList = append(newList, p)
		}
	}
	return newList
}

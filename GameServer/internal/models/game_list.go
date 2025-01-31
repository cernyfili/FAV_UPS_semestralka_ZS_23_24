package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/utils/errorHandeling"
	"sync"
)

// region DATA STRUCTURES
type GameList struct {
	list *List
}

var (
	instanceGL *GameList
	onceGL     sync.Once
)

//endregion

//region FUNCTIONS

func GetInstanceGameList() *GameList {
	onceGL.Do(func() {
		instanceGL = &GameList{
			list: CreateList(),
		}
	})
	return instanceGL
}

// player is in game
func (gl *GameList) PlayerIsInGame(player *Player) bool {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	for _, v := range gl.list.data {
		game, ok := v.(*Game)
		if !ok {
			panic("item is not a game")
			return false
		}
		for _, p := range game.playersGameDataArr {
			if p.Player == player {
				return true
			}
		}
	}
	return false
}

func (gl *GameList) AddItem(game *Game) (int, error) {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	key, err := gl.list.AddItem(game)
	if err != nil {
		errorHandeling.PrintError(err)
		return -1, err
	}
	gameItem := gl.list.GetItemWithoutLock(key)
	if gameItem == nil {
		return -1, fmt.Errorf("game not found")
	}
	game = gameItem.(*Game)
	game.gameID = key

	return key, nil
}

// has item
func (gl *GameList) HasItemName(gameName string) bool {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	game, err := gl.getItemByName(gameName)
	return err == nil && game != nil
}

func (gl *GameList) getItemByName(gameName string) (*Game, error) {
	for _, v := range gl.list.data {
		game, ok := v.(*Game)
		if !ok {
			return nil, fmt.Errorf("item is not a game")
		}
		if game.name == gameName {
			return game, nil
		}
	}
	return nil, fmt.Errorf("game not found")
}

// remove item
func (gl *GameList) RemoveItem(game *Game) error {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	gameKey := game.gameID

	err := gl.list.RemoveItem(gameKey)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}

//func (gl *GameList) GetItem(key int) (*Game, error) {
//	gl.list.mutex.Lock()
//	defer gl.list.mutex.Unlock()
//
//	gameItem := gl.list.GetItemWithoutLock(key)
//	if gameItem == nil {
//		return nil, fmt.Errorf("gameItem not found")
//	}
//	game, ok := gameItem.(*Game)
//	if !ok {
//		return nil, fmt.Errorf("gameItem is not a game")
//	}
//	return game, nil
//}

// Remove Player from Game
func (gl *GameList) RemovePlayerFromGame(player *Player) error {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	if player == nil {
		err := fmt.Errorf("player is nil")
		errorHandeling.PrintError(err)
		return err
	}

	game := gl.getPlayersGame(player)
	if game == nil {
		return nil
	}

	err := game.RemovePlayer(player)
	if err != nil {
		err = fmt.Errorf("cannot remove player %w", err)
		errorHandeling.PrintError(err)
		panic(err)
	}

	return nil
}

// get item by game name
func (gl *GameList) GetItemByName(name string) (*Game, error) {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	for _, v := range gl.list.data {
		game, ok := v.(*Game)
		if !ok {
			return nil, fmt.Errorf("item is not a game")
		}
		if game.name == name {
			return game, nil
		}
	}
	return nil, fmt.Errorf("game not found")
}

// Has Item in list
func (gl *GameList) HasValue(game *Game) bool {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	return gl.list.HasValue(game)
}

// GetValuesArray
func (gl *GameList) GetValuesArray() []*Game {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	var values []*Game
	for _, v := range gl.list.data {
		values = append(values, v.(*Game))
	}

	return values
}

// GetCreatedGameList
func (gl *GameList) GetCreatedGameList() []*Game {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	var values []*Game
	for _, v := range gl.list.data {
		game, ok := v.(*Game)
		if !ok {
			panic("item is not a game")
			return nil
		}
		if game.gameStateValue == Created {
			values = append(values, game)
		}
	}

	return values
}

// GetPlayersGame returns the game of the player
func (gl *GameList) getPlayersGame(player *Player) *Game {
	for _, v := range gl.list.data {
		game, ok := v.(*Game)
		if !ok {
			panic("item is not a game")
			return nil
		}
		for _, p := range game.playersGameDataArr {
			if p.Player == player {
				return game
			}
		}
	}
	return nil
}

func (gl *GameList) GetPlayersGame(player *Player) *Game {
	logger.Log.Debugln("GetPlayersGame before lock")
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	return gl.getPlayersGame(player)
}

//endregion

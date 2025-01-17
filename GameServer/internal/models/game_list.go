package models

import (
	"fmt"
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

func (gl *GameList) AddItem(game *Game) (int, error) {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	key, err := gl.list.AddItem(game)
	gameItem, err := gl.list.GetItemWithoutLock(key)
	if err != nil {
		errorHandeling.PrintError(err)
		return -1, err
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

func (gl *GameList) GetItem(key int) (*Game, error) {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	item, err := gl.list.GetItemWithoutLock(key)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, err
	}
	game, ok := item.(*Game)
	if !ok {
		return nil, fmt.Errorf("item is not a game")
	}
	return game, nil
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

// GetPlayersGame returns the game of the player
func (gl *GameList) GetPlayersGame(player *Player) *Game {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

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

//endregion

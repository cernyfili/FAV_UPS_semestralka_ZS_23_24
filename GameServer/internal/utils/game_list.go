package utils

import (
	"fmt"
	"sync"
)

type GameList struct {
	list *List
}

var (
	instanceGL *GameList
	onceGL     sync.Once
)

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
	gameItem, err := gl.list.GetItem(key)
	if err != nil {
		return -1, err
	}
	game = gameItem.(*Game)
	game.gameID = key

	return key, nil
}

func (gl *GameList) GetItem(key int) (*Game, error) {
	gl.list.mutex.Lock()
	defer gl.list.mutex.Unlock()

	item, err := gl.list.GetItem(key)
	if err != nil {
		return nil, err
	}
	game, ok := item.(*Game)
	if !ok {
		return nil, fmt.Errorf("item is not a game")
	}
	return game, nil
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

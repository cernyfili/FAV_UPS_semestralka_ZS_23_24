package data_structure

import "errors"

type GameList struct {
	list *List
}

func CreateGameList() *GameList {
	return &GameList{
		list: CreateList(),
	}
}

func (gl *GameList) AddItem(game *Game) (int, error) {
	key, err := gl.list.AddItem(game)
	gameItem, err := gl.list.GetItem(key)
	if err != nil {
		return -1, err
	}
	game = gameItem.(*Game)
	game.GameID = key

	return key, nil
}

func (gl *GameList) GetItem(key int) (*Game, error) {
	item, err := gl.list.GetItem(key)
	if err != nil {
		return nil, err
	}
	game, ok := item.(*Game)
	if !ok {
		return nil, errors.New("item is not a game")
	}
	return game, nil
}

// Has Item in list
func (gl *GameList) HasValue(game *Game) bool {
	return gl.list.HasValue(game)
}

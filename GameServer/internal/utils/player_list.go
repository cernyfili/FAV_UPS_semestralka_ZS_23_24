package utils

import (
	"fmt"
	"sync"
)

type PlayerList struct {
	list *List
}

var (
	instancePL *PlayerList
	oncePL     sync.Once
)

func GetInstancePlayerList() *PlayerList {
	oncePL.Do(func() {
		instancePL = &PlayerList{
			list: CreateList(),
		}
	})
	return instancePL
}

func (pl *PlayerList) AddItem(key string, player *Player) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	err := pl.list.AddItemKey(key, player)
	if err != nil {
		return err
	}
	return nil
}

func (pl *PlayerList) GetItem(key string) (*Player, error) {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	item, err := pl.list.GetItem(key)
	if err != nil {
		return nil, err
	}
	player, ok := item.(*Player)
	if !ok {
		return nil, fmt.Errorf("item is not a player")
	}
	return player, nil
}

// Has Item in list
func (pl *PlayerList) HasValue(player *Player) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	return pl.list.HasValue(player)
}

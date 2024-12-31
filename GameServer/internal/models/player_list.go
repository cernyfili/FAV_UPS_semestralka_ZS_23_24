package models

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

func (pl *PlayerList) AddItem(player *Player) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	key := player.GetNickname()

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
		return nil, fmt.Errorf("item is not a Player")
	}
	return player, nil
}

func (pl *PlayerList) HasItem(key string) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	_, err := pl.list.GetItem(key)
	if err != nil {
		return false
	}

	return true
}

// Has Item in list
func (pl *PlayerList) HasValue(player *Player) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	return pl.list.HasValue(player)
}

func (pl *PlayerList) RemoveItem(player *Player) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	key := player.GetNickname()

	err := pl.list.RemoveItem(key)
	if err != nil {
		return err
	}
	return nil

}

// GetValuesArray
func (pl *PlayerList) GetValuesArray() []*Player {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	var values []*Player
	for _, v := range pl.list.data {
		values = append(values, v.(*Player))
	}

	return values
}

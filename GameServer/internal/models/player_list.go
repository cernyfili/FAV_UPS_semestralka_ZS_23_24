package models

import (
	"fmt"
	"gameserver/internal/utils/errorHandeling"
	"net"
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

// get player list where player is in state in argument
func (pl *PlayerList) GetActivePlayersInState(state string) []*Player {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	var players []*Player
	for _, v := range pl.list.data {
		player := v.(*Player)
		if player.GetCurrentStateName() == state && player.IsConnected() {
			players = append(players, player)
		}
	}

	return players
}

func (pl *PlayerList) AddItem(player *Player) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	key := player.GetNickname()

	err := pl.list.AddItemKey(key, player)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	return nil
}

func (pl *PlayerList) GetItem(key string) (*Player, error) {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}

	item := pl.list.GetItemWithoutLock(key)

	player, ok := item.(*Player)
	if !ok {
		return nil, fmt.Errorf("item is not a Player")
	}
	return player, nil
}

func (pl *PlayerList) HasItemName(playerName string) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	item := pl.list.GetItemWithoutLock(playerName)
	if item == nil {
		return false
	}

	return true
}

// GetPlayerByConnection
func (pl *PlayerList) GetPlayerByConnection(connection net.Conn) *Player {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	for _, v := range pl.list.data {
		player := v.(*Player)
		if player.GetConnectionInfo().Connection == connection {
			return player
		}
	}

	return nil
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

	if player == nil {
		return fmt.Errorf("player is nil")
	}

	key := player.GetNickname()

	err := pl.list.RemoveItem(key)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	return nil

}

// GetValuesArrayWithoutOnePlayer
func (pl *PlayerList) GetValuesArrayWithoutOnePlayer(player *Player) []*Player {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	var values []*Player
	for _, v := range pl.list.data {
		if v.(*Player) == player {
			continue
		}
		values = append(values, v.(*Player))
	}

	return values
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

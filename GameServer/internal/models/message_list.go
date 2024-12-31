package models

import (
	"fmt"
	"sync"
)

type MessageList struct {
	list *List
}

var (
	instanceML *MessageList
	onceML     sync.Once
)

func GetInstanceMessageList() *MessageList {
	onceML.Do(func() {
		instanceML = &MessageList{
			list: CreateList(),
		}
	})
	return instanceML
}

func (pl *MessageList) AddItem(message Message) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	key := message.PlayerNickname + message.TimeStamp
	err := pl.list.AddItemKey(key, message)
	if err != nil {
		return err
	}
	return nil
}

func (pl *MessageList) GetItem(nickname string, timestamp string) (Message, error) {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	key := nickname + timestamp
	item, err := pl.list.GetItem(key)
	if err != nil {
		return Message{}, err
	}
	message, ok := item.(Message)
	if !ok {
		return Message{}, fmt.Errorf("item is not a message")
	}

	return message, nil
}

// Has Item in list
func (pl *MessageList) HasValue(player *Player) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	return pl.list.HasValue(player)
}

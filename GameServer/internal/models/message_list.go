package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/utils/errorHandeling"
)

type MessageType int

// enum which is send or received
const (
	Send MessageType = iota
	Received
)

type MessageList struct {
	list     *List
	typeMess MessageType
}

func CreateMessageList(typeMess MessageType) *MessageList {
	return &MessageList{
		list:     CreateList(),
		typeMess: typeMess,
	}
}

func (pl *MessageList) AddItem(message Message) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	logger.Log.Infof("Message type %v -> %s", pl.typeMess, message.String())

	key := message.PlayerNickname + message.TimeStamp
	err := pl.list.AddItemKey(key, message)
	if err != nil {
		errorHandeling.PrintError(err)
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
		errorHandeling.PrintError(err)
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

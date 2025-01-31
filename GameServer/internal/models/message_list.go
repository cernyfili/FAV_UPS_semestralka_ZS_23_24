package models

import (
	"fmt"
	"gameserver/internal/logger"
)

type MessageType int

func (mt MessageType) String() string {
	switch mt {
	case Send:
		return "Send"
	case Received:
		return "Received"
	default:
		return fmt.Sprintf("Unknown(%d)", mt)
	}
}

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

func (pl *MessageList) AddItem(message Message) {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	if message.CommandID != 50 && message.CommandID != 60 {
		//todo change
		logger.Log.Infof("Message type %v:\n%s", pl.typeMess.String(), message.String())
	}
	//logger.Log.Infof("Message type %v:\n%s", pl.typeMess.String(), message.String())

	commandIDstr := fmt.Sprintf("%d", message.CommandID)

	key := message.PlayerNickname + message.TimeStamp + commandIDstr
	err := pl.list.AddItemKey(key, message)
	if err != nil {
		logger.Log.Errorf("Error adding message to list: %v", err)
	}
}

func (pl *MessageList) GetMessagesByPlayer(nickname string) ([]Message, error) {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	var messages []Message
	for _, item := range pl.list.data {
		message, ok := item.(Message)
		if !ok {
			return nil, fmt.Errorf("item is not a message")
		}
		if message.PlayerNickname == nickname {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

// Has Item in list
func (pl *MessageList) HasValue(player *Player) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	return pl.list.HasValue(player)
}

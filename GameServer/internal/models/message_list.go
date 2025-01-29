package models

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/utils/errorHandeling"
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

func (pl *MessageList) AddItem(message Message) error {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	logger.Log.Infof("Message type %v:\n%s", pl.typeMess.String(), message.String())

	commandIDstr := fmt.Sprintf("%d", message.CommandID)

	key := message.PlayerNickname + message.TimeStamp + commandIDstr
	err := pl.list.AddItemKey(key, message)
	if err != nil {
		errorHandeling.PrintError(err)
		return err
	}
	return nil
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

//func (pl *MessageList) GetItemWithoutLock(nickname string, timestamp string) (Message, error) {
//	pl.list.mutex.Lock()
//	defer pl.list.mutex.Unlock()
//
//	key := nickname + timestamp
//	item, err := pl.list.GetItemWithoutLock(key)
//	if err != nil {
//		errorHandeling.PrintError(err)
//		return Message{}, err
//	}
//	message, ok := item.(Message)
//	if !ok {
//		return Message{}, fmt.Errorf("item is not a message")
//	}
//
//	return message, nil
//}

// Has Item in list
func (pl *MessageList) HasValue(player *Player) bool {
	pl.list.mutex.Lock()
	defer pl.list.mutex.Unlock()

	return pl.list.HasValue(player)
}

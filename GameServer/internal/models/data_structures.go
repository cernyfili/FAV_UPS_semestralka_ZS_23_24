package models

import (
	"fmt"
	"gameserver/internal/utils/constants"
	"net"
	"strings"
	"time"
)

type ConnectionInfo struct {
	Connection net.Conn
	TimeStamp  string
}

type MessageInfo struct {
	ConnectionInfo ConnectionInfo
	PlayerNickname string
}

type MessageHeader struct {
	Signature      [len(constants.CMessageSignature)]byte
	CommandID      [2]byte
	TimeStamp      [26]byte
	PlayerNickname string // {karel_1}
}

type Message struct {
	Signature      string
	CommandID      int
	TimeStamp      string
	PlayerNickname string
	Parameters     []constants.Params
}

//endregion

// region FUNCTIONS CONSTRUCTORS
func CreateMessage(playerName string, commandID int, params []constants.Params) Message {
	return Message{
		Signature: constants.CMessageSignature,
		CommandID: commandID,
		//current time
		TimeStamp:      time.Now().Format(constants.CMessageTimeFormat),
		PlayerNickname: playerName,
		Parameters:     params,
	}
}

// func to check if is valid message name
func IsValidName(name string) bool {
	// if name is none
	if name == "" {
		return false
	}

	// if name is not numbers and letters
	if !constants.IsAlphaNumeric(name) {
		return false
	}

	return len(name) >= constants.CMessageNameMinChars && len(name) <= constants.CMessageNameMaxChars
}

func CreateParams(names []string, values []string) ([]constants.Params, error) {
	var params []constants.Params
	if len(names) != len(values) {
		return params, fmt.Errorf("error creating params")
	}

	for i := 0; i < len(names); i++ {
		param := constants.Params{
			Name:  names[i],
			Value: values[i],
		}
		params = append(params, param)
	}

	return params, nil
}

func (m *Message) String() string {
	paramsStringList := make([]string, 0)
	for _, param := range m.Parameters {
		paramsStringList = append(paramsStringList, param.Name+": "+param.Value)
	}
	return fmt.Sprintf("Signature: %s\nCommand: %d:%s\nTimeStamp: %s\nPlayerNickname: %s\nParameters: %s",
		m.Signature, m.CommandID, constants.GetCommandName(m.CommandID), m.TimeStamp, m.PlayerNickname, strings.Join(paramsStringList, ", "))
}

//endregion

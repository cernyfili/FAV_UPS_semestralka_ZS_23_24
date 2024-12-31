package models

import (
	"fmt"
	"gameserver/internal/utils"
	"net"
)

type ConnectionInfo struct {
	Connection net.Conn
	TimeStamp  string
}

type NetworkResponseInfo struct {
	ConnectionInfo ConnectionInfo
	PlayerNickname string
}

type MessageHeader struct {
	Signature      [6]byte
	CommandID      [2]byte
	TimeStamp      [32]byte
	PlayerNickname string // {karel_1}
}

type Message struct {
	Signature      string
	CommandID      int
	TimeStamp      string
	PlayerNickname string
	Parameters     []utils.Params
}

//endregion

// region FUNCTIONS CONSTRUCTORS
func CreateResponseMessage(responseInfo NetworkResponseInfo, commandID int, params []utils.Params) Message {
	return Message{
		Signature:      utils.CMessageSignature,
		CommandID:      commandID,
		TimeStamp:      responseInfo.ConnectionInfo.TimeStamp, //original so client can match response
		PlayerNickname: responseInfo.PlayerNickname,
		Parameters:     params,
	}
}

func CreateParams(names []string, values []string) ([]utils.Params, error) {
	var params []utils.Params
	if len(names) != len(values) {
		return params, fmt.Errorf("error creating params")
	}

	for i := 0; i < len(names); i++ {
		param := utils.Params{
			Name:  names[i],
			Value: values[i],
		}
		params = append(params, param)
	}

	return params, nil
}

//endregion

package utils

import (
	"fmt"
	"net"
)

type Params struct {
	Name  string
	Value string
}

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
	CommandID      byte
	TimeStamp      [32]byte
	PlayerNickname string // {karel_1}
}

type Message struct {
	Signature      string
	CommandID      int
	TimeStamp      string
	PlayerNickname string
	Parameters     []Params
}

func CreateResponseMessage(responseInfo NetworkResponseInfo, commandID int, params []Params) Message {
	return Message{
		Signature:      CMessageSignature,
		CommandID:      commandID,
		TimeStamp:      responseInfo.ConnectionInfo.TimeStamp, //original so client can match response
		PlayerNickname: responseInfo.PlayerNickname,
		Parameters:     params,
	}
}

func CreateParams(names []string, values []string) ([]Params, error) {
	var params []Params
	if len(names) != len(values) {
		return params, fmt.Errorf("error creating params")
	}

	for i := 0; i < len(names); i++ {
		param := Params{
			Name:  names[i],
			Value: values[i],
		}
		params = append(params, param)
	}

	return params, nil
}

package utils

import (
	"net"
)

type PlayerState int

/*const (
	Running PlayerState = iota
	Created
	Ended
)*/

// Turn represents a player's turn with their ID and score.
type Turn struct {
	PlayerID int
	Score    int
}

type ConnectionInfo struct {
	Connection net.Conn
	TimeStamp  string
}

type NetworkResponseInfo struct {
	ConnectionInfo ConnectionInfo
	PlayerNickname string
}

// Player represents a player with a unique ID and nickname.
type Player struct {
	Nickname       string
	Game           *Game
	IsConnected    bool
	ConnectionInfo ConnectionInfo
}

type MessageHeader struct { //todo add timestemp for error when client doesnt get response
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
	Parameters     string
}

func CreateResponseMessage(responseInfo NetworkResponseInfo, commandID int, params string) Message {
	return Message{
		Signature:      CMessageSignature,
		CommandID:      commandID,
		TimeStamp:      responseInfo.ConnectionInfo.TimeStamp, //original so client can match response
		PlayerNickname: responseInfo.PlayerNickname,
		Parameters:     params,
	}
}

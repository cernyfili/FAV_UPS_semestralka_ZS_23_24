package internal

import (
	"fmt"
	"gameserver/internal/utils"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
) //todo predelat bez pouziti regex

var gRegexPatterns = map[string]string{
	"parseGameIDPlayerID":  `^\{"gameID":"(\d+)","playerID":"(\d+)"\}$`,
	"parsePlayerLoginArgs": `^\{"nickname":"([A-Za-z0-9_\\-]+)"\}$`,
	"parseCreateGameArgs":  `^\{"name":"([A-Za-z0-9_\\-]+)","maxPlayers":"(\d+)"\}$`,
	"parsePlayerID":        `^\{"playerID":"(\d+)"\}$`,
}

type Brackets struct {
	Opening string
	Closing string
}

var cPlayerNicknameBrackets = Brackets{Opening: "{", Closing: "}"}

func ParseMessage(input string) (utils.Message, error) {
	messageData := utils.MessageHeader{}
	signatureSize := int(unsafe.Sizeof(messageData.Signature))
	commandIDSize := int(unsafe.Sizeof(messageData.CommandID))
	timeStempSize := int(unsafe.Sizeof(messageData.TimeStamp))
	minSizeMessage := signatureSize + commandIDSize + timeStempSize

	if len(input) < minSizeMessage {
		return utils.Message{}, fmt.Errorf("invalid input format")
	}

	start := 0

	//Read Signature
	signature := input[start : start+signatureSize]
	start += signatureSize
	start++

	//Read Command ID
	commandID := input[start : start+commandIDSize]
	start += commandIDSize
	start++

	//Read TimeStemp
	timeStamp := input[start : start+timeStempSize]
	start += timeStempSize
	start++

	//Read Nickname
	playerNickname, playerNicknameSize, err := parseMessagePlayerID(input[start:])
	if err != nil {
		return utils.Message{}, fmt.Errorf("error parsing player ID: %v", err)
	}
	start += playerNicknameSize
	start++

	parameters := input[start:]

	//convert values
	commandIDint, err := strconv.ParseUint(commandID, 10, commandIDSize)
	if err != nil {
		return utils.Message{}, fmt.Errorf("error parsing command ID: %v", err)
	}

	return utils.Message{
		Signature:      signature, // You can populate this with the actual signature values
		CommandID:      int(commandIDint),
		TimeStamp:      timeStamp,
		PlayerNickname: playerNickname,
		Parameters:     parameters,
	}, nil
}

func parseMessagePlayerID(input string) (string, int, error) {
	playerNickname := ""
	playerNicknameSize := 0

	if len(input) == 0 {
		return playerNickname, playerNicknameSize, fmt.Errorf("invalid player ID format")
	}

	opening := cPlayerNicknameBrackets.Opening
	closing := cPlayerNicknameBrackets.Closing

	playerNickname, err := extractSubstring(input, opening, closing)
	if err != nil {
		return playerNickname, playerNicknameSize, fmt.Errorf("invalid player ID format")
	}

	playerNicknameSize = len(playerNickname) + len(opening) + len(closing)

	return playerNickname, playerNicknameSize, nil
}

func extractSubstring(input, opening, closing string) (string, error) {
	startIndex := strings.Index(input, opening)
	if startIndex == -1 {
		return "", fmt.Errorf("opening character not found")
	}
	endIndex := strings.Index(input[startIndex:], closing)
	if endIndex == -1 {
		return "", fmt.Errorf("closing character not found")
	}
	return input[startIndex+len(opening) : startIndex+endIndex], nil
}

// region PRIVATE FUNCTIONS
func parseGameIDPlayerID(str string) (int, int, error) {
	gameID := -1
	playerID := -1

	// Define the regular expression pattern to match the args
	pattern := gRegexPatterns["parseGameIDPlayerID"]

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the matches in the string
	matches := regex.FindStringSubmatch(str)

	// Check if there is a match and extract the name and maxPlayers
	if len(matches) != 3 {
		return gameID, playerID, fmt.Errorf("invalid args format")
	}

	gameID, err := strconv.Atoi(matches[1])
	if err != nil {
		return gameID, playerID, fmt.Errorf("invalid maxPlayers format")
	}

	playerID, err = strconv.Atoi(matches[2])
	if err != nil {
		return gameID, playerID, fmt.Errorf("invalid maxPlayers format")
	}

	return gameID, playerID, nil
}

func parsePlayerID(str string) (int, error) {
	playerID := -1

	// Define the regular expression pattern to match the args
	pattern := gRegexPatterns["parsePlayerID"]

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the matches in the string
	matches := regex.FindStringSubmatch(str)

	// Check if there is a match and extract the nickname
	if len(matches) != 2 {
		return playerID, fmt.Errorf("invalid args format")
	}

	playerID, err := strconv.Atoi(matches[1])
	if err != nil {
		return playerID, fmt.Errorf("invalid playerID format")
	}

	return playerID, nil
}

//endregion

// region PUBLIC FUNCTIONS
func ParsePlayerLoginArgs(str string) (string, error) {
	nickname := ""

	// Define the regular expression pattern to match the args
	pattern := gRegexPatterns["parsePlayerLoginArgs"]

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the matches in the string
	matches := regex.FindStringSubmatch(str)

	// Check if there is a match and extract the nickname
	if len(matches) != 2 {
		return nickname, fmt.Errorf("invalid args format")
	}

	nickname = matches[1]
	return nickname, nil
}

func ParseCreateGameArgs(str string) (string, int, error) {
	name := ""
	maxPlayers := -1

	// Define the regular expression pattern to match the args
	pattern := gRegexPatterns["parseCreateGameArgs"]

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the matches in the string
	matches := regex.FindStringSubmatch(str)

	// Check if there is a match and extract the name and maxPlayers
	if len(matches) != 3 {
		return name, maxPlayers, fmt.Errorf("invalid args format")
	}

	name = matches[1]
	maxPlayers, err := strconv.Atoi(matches[2])
	if err != nil {
		return name, maxPlayers, fmt.Errorf("invalid maxPlayers format")
	}

	return name, maxPlayers, nil
}

func ParseJoinGameArgs(str string) (int, int, error) {
	return parseGameIDPlayerID(str)
}

func ParseStartGameArgs(str string) (int, int, error) {
	return parseGameIDPlayerID(str)
}

func ParseRollDiceArgs(str string) (int, error) {
	return parsePlayerID(str)
}

func ParsePlayerLogoutArgs(str string) (int, error) {
	return parsePlayerID(str)
}

func ParseEndGameArgs(str string) (int, int, error) {
	return parseGameIDPlayerID(str)
}

//endregion

func ConvertMessageToNetworkString(message utils.Message) (string, error) {
	networkString := ""

	messageData := utils.MessageHeader{}
	signatureSize := int(unsafe.Sizeof(messageData.Signature))
	commandIDSize := int(unsafe.Sizeof(messageData.CommandID))
	timeStempSize := int(unsafe.Sizeof(messageData.TimeStamp))

	//Signature
	if len(message.Signature) != signatureSize {
		return networkString, fmt.Errorf("invalid signature")
	}
	networkString += message.Signature

	//Command ID
	commandIDStr := fmt.Sprintf("%d", message.CommandID)
	if len(commandIDStr) != commandIDSize {
		return networkString, fmt.Errorf("invalid command ID")
	}
	networkString += commandIDStr

	//TimeStemp
	if len(message.TimeStamp) != timeStempSize {
		return networkString, fmt.Errorf("invalid time stamp")
	}
	networkString += message.TimeStamp

	//PlayerNickname
	networkString += convertPlayerNicknameToNetworkString(message.PlayerNickname)

	return networkString, nil
}

func convertPlayerNicknameToNetworkString(nickname string) string {
	networkStr := cPlayerNicknameBrackets.Opening
	networkStr += nickname
	networkStr += cPlayerNicknameBrackets.Closing

	return networkStr
}

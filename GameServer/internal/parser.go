package internal

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/utils"
	"gameserver/internal/utils/errorHandeling"
	"strconv"
	"strings"
	"unsafe"
) //todo predelat bez pouziti regex

// region CONSTANTS
/*var gRegexPatterns = map[string]string{
	"parseGameIDPlayerID":  `^\{"gameID":"(\d+)","playerID":"(\d+)"\}$`,
	"parsePlayerLoginArgs": `^\{"nickname":"([A-Za-z0-9_\\-]+)"\}$`,
	"parseCreateGameArgs":  `^\{"name":"([A-Za-z0-9_\\-]+)","maxPlayers":"(\d+)"\}$`,
	"parsePlayerID":        `^\{"playerID":"(\d+)"\}$`,
}*/

//endregion

// region DATA STRUCTURES
type Brackets struct {
	Opening string
	Closing string
}

var (
	cParamsBrackets          = Brackets{Opening: "{", Closing: "}"}
	cArrayBrackets           = Brackets{Opening: "[", Closing: "]"}
	cParamsDelimiter         = ","
	cParamsKeyValueDelimiter = ":"
	cParamsWrapper           = "\""
)

//endregion

// region FUNCTIONS PARSE
func ParseMessage(input string) (models.Message, error) {
	logger.Log.Debug("Parsing message: %v", input)

	//remove end delimiter
	if len(input) > 0 && input[len(input)-1] == utils.CMessageEndDelimiter {
		input = input[:len(input)-1]
	}

	messageData := models.MessageHeader{}
	signatureSize := int(unsafe.Sizeof(messageData.Signature))
	commandIDSize := int(unsafe.Sizeof(messageData.CommandID))
	timeStempSize := int(unsafe.Sizeof(messageData.TimeStamp))
	minSizeMessage := signatureSize + commandIDSize + timeStempSize

	if len(input) < minSizeMessage {
		return models.Message{}, fmt.Errorf("invalid input format")
	}

	start := 0

	//Read Signature
	signature := input[start : start+signatureSize]
	start += signatureSize

	//Read Command ID
	commandID := input[start : start+commandIDSize]
	start += commandIDSize

	//Read TimeStemp
	timeStamp := input[start : start+timeStempSize]
	start += timeStempSize

	//Read nickname
	playerNickname, playerNicknameSize, err := parseMessagePlayerID(input[start:])
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, fmt.Errorf("error parsing player ID: %v", err)
	}
	start += playerNicknameSize

	//Read Parameters
	parametersStr := input[start:]

	params, err := parseParamsStr(parametersStr)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, fmt.Errorf("error parsing params: %v", err)
	}

	//convert values
	commandIDint, err := strconv.ParseUint(commandID, 10, commandIDSize)
	if err != nil {
		errorHandeling.PrintError(err)
		return models.Message{}, fmt.Errorf("error parsing command ID: %v", err)
	}

	return models.Message{
		Signature:      signature, // You can populate this with the actual signature values
		CommandID:      int(commandIDint),
		TimeStamp:      timeStamp,
		PlayerNickname: playerNickname,
		Parameters:     params,
	}, nil
}

func ParseParamValueArray(value string) ([]string, error) {
	var valueArray []string

	if len(value) == 0 {
		return valueArray, nil
	}

	if value[0] != cArrayBrackets.Opening[0] {
		return valueArray, fmt.Errorf("invalid valueArray format")
	}

	if value[len(value)-1] != cArrayBrackets.Closing[0] {
		return valueArray, fmt.Errorf("invalid valueArray format")
	}

	value = value[1 : len(value)-1]

	valuesStr := strings.Split(value, cParamsDelimiter)

	for _, valueStr := range valuesStr {
		valueArray = append(valueArray, valueStr)
	}

	return valueArray, nil
}

func parseParamsStr(paramsString string) ([]utils.Params, error) {
	var paramArray []utils.Params

	if len(paramsString) == 0 {
		return paramArray, nil
	}

	if paramsString[0] != cParamsBrackets.Opening[0] {
		return paramArray, fmt.Errorf("invalid paramArray format")
	}

	if paramsString[len(paramsString)-1] != cParamsBrackets.Closing[0] {
		return paramArray, fmt.Errorf("invalid paramArray format")
	}

	paramsString = paramsString[1 : len(paramsString)-1]

	paramsStr := strings.Split(paramsString, cParamsDelimiter)

	for _, paramStr := range paramsStr {
		parameter, err := parseParam(paramStr)
		if err != nil {
			errorHandeling.PrintError(err)
			return paramArray, fmt.Errorf("invalid paramArray format")
		}
		paramArray = append(paramArray, parameter)
	}

	return paramArray, nil
}

func parseParam(str string) (utils.Params, error) {
	parameter := utils.Params{}

	if len(str) == 0 {
		return parameter, nil
	}

	paramKeyValue := strings.Split(str, cParamsDelimiter)

	if len(paramKeyValue) != 2 {
		return parameter, fmt.Errorf("invalid param format")
	}

	parameter.Name = paramKeyValue[0]
	if parameter.Name[0] != cParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}

	if parameter.Name[len(parameter.Name)-1] != cParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}
	parameter.Name = parameter.Name[1 : len(parameter.Name)-1]

	parameter.Value = paramKeyValue[1]
	if parameter.Value[0] != cParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}
	if parameter.Value[len(parameter.Value)-1] != cParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}
	parameter.Value = parameter.Value[1 : len(parameter.Value)-1]

	return parameter, nil
}

func parseMessagePlayerID(input string) (string, int, error) {
	playerNickname := ""
	playerNicknameSize := 0

	if len(input) == 0 {
		return playerNickname, playerNicknameSize, fmt.Errorf("invalid player ID format")
	}

	opening := cParamsBrackets.Opening
	closing := cParamsBrackets.Closing

	playerNickname, err := extractSubstring(input, opening, closing)
	if err != nil {
		errorHandeling.PrintError(err)
		return playerNickname, playerNicknameSize, fmt.Errorf("invalid player ID format")
	}

	playerNicknameSize = len(playerNickname) + len(opening) + len(closing)

	return playerNickname, playerNicknameSize, nil
}

//endregion

// region FUNCTIONS UTILS
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

func getArrayElementStr(name string, value string) string {
	returnStr := cParamsWrapper + name + cParamsWrapper + cParamsKeyValueDelimiter + cParamsWrapper + value + cParamsWrapper + cParamsDelimiter
	return returnStr
}

//endregion

// region FUNCTIONS CONVERT
func ConvertMessageToNetworkString(message models.Message) (string, error) {
	networkString := ""

	messageData := models.MessageHeader{}
	signatureSize := int(unsafe.Sizeof(messageData.Signature))
	commandIDSize := int(unsafe.Sizeof(messageData.CommandID))
	timeStempSize := int(unsafe.Sizeof(messageData.TimeStamp))

	//Signature
	if len(message.Signature) != signatureSize {
		return networkString, fmt.Errorf("invalid signature")
	}
	networkString += message.Signature

	//Command ID
	commandIDStr := fmt.Sprintf("%02d", message.CommandID)
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

	//Parameters
	paramsStr, err := convertParamsToNetworkString(message.Parameters)
	if err != nil {
		errorHandeling.PrintError(err)
		return networkString, err
	}
	networkString += paramsStr

	return networkString, nil
}

func ConvertGameListToNetworkString(array []*models.Game) string {
	arrayStr := ""

	arrayStr += cArrayBrackets.Opening

	for _, game := range array {

		gameName := game.GetName()
		maxPlayers := fmt.Sprintf("%d", game.GetMaxPlayers())
		playersCount := fmt.Sprintf("%d", game.GetPlayersCount())

		arrayStr += cParamsBrackets.Opening
		arrayStr += getArrayElementStr("name", gameName)
		arrayStr += getArrayElementStr("maxPlayers", maxPlayers)
		arrayStr += getArrayElementStr("playersCount", playersCount)
		arrayStr += cParamsBrackets.Closing + cParamsDelimiter
	}

	arrayStr += cArrayBrackets.Closing

	return arrayStr
}

func ConvertPlayerListToNetworkString(array []*models.Player) string {
	arrayStr := ""

	arrayStr += cArrayBrackets.Opening

	for _, player := range array {

		playerNickname := player.GetNickname()

		arrayStr += cParamsBrackets.Opening
		arrayStr += cParamsWrapper + "nickname" + cParamsWrapper + cParamsKeyValueDelimiter + cParamsWrapper + playerNickname + cParamsWrapper
		arrayStr += cParamsBrackets.Closing + cParamsDelimiter
	}

	arrayStr += cArrayBrackets.Closing

	return arrayStr
}

func ConvertGameDataToNetworkString(data models.GameData) string {
	arrayStr := ""

	arrayStr += cArrayBrackets.Opening

	for _, playerGameData := range data.PlayerGameDataArr {

		nickname := playerGameData.Player.GetNickname()
		isConnected := fmt.Sprintf("%t", playerGameData.Player.IsConnected())
		score := fmt.Sprintf("%d", playerGameData.Score)
		isTurnBool := playerGameData.Player == data.TurnPlayer
		isTurn := fmt.Sprintf("%t", isTurnBool)

		arrayStr += cParamsBrackets.Opening
		arrayStr += getArrayElementStr("nickname", nickname)
		arrayStr += getArrayElementStr("isConnected", isConnected)
		arrayStr += getArrayElementStr("score", score)
		arrayStr += getArrayElementStr("isTurn", isTurn)

		arrayStr += cParamsBrackets.Closing + cParamsDelimiter
	}

	arrayStr += cArrayBrackets.Closing

	return arrayStr
}

func ConvertCubeValuesToNetworkString(values []int) string {
	arrayStr := ""

	arrayStr += cArrayBrackets.Opening

	for _, value := range values {

		valueStr := fmt.Sprintf("%d", value)

		arrayStr += cParamsBrackets.Opening
		arrayStr += getArrayElementStr("value", valueStr)
		arrayStr += cParamsBrackets.Closing + cParamsDelimiter
	}

	arrayStr += cArrayBrackets.Closing

	return arrayStr
}

func convertPlayerNicknameToNetworkString(nickname string) string {
	networkStr := ""
	networkStr += cParamsBrackets.Opening
	networkStr += nickname
	networkStr += cParamsBrackets.Closing

	return networkStr
}

func convertParamsToNetworkString(params []utils.Params) (string, error) {
	length := len(params)

	paramsStr := ""
	paramsStr += cParamsBrackets.Opening
	for i := 0; i < length; i++ {
		paramsStr += cParamsWrapper + params[i].Name + cParamsWrapper + cParamsKeyValueDelimiter + cParamsWrapper + params[i].Value + cParamsWrapper + cParamsDelimiter
	}
	paramsStr += cParamsBrackets.Closing

	return paramsStr, nil
}

//endregion

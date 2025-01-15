package parser

import (
	"fmt"
	"gameserver/internal/logger"
	"gameserver/internal/models"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
) //todo predelat bez pouziti regex

//region FUNCTIONS PARSE

func ParseMessage(input string) (models.Message, error) {
	logger.Log.Debug("Parsing message: %v", input)

	parts := strings.Split(input, constants.CMessageEndDelimiter)
	// if len is not 2 and second part is not empty
	if len(parts) != 2 && len(parts[1]) != 0 {
		err := fmt.Errorf("invalid input format")
		errorHandeling.PrintError(err)
		return models.Message{}, err
	}

	input = parts[0]

	if len(input) == 0 {
		err := fmt.Errorf("empty input")
		errorHandeling.PrintError(err)
		return models.Message{}, err
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
	commandIDint, err := strconv.ParseUint(commandID, 10, 64)
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

/*
ConvertParamClientCreateGame

	checks if the parameters are correct and converts them to the correct type
*/
func ConvertParamClientCreateGame(params []constants.Params, names []string) (string, int, error) {
	gameName := ""
	maxPlayers := -1
	var err error

	if len(params) != len(names) {
		return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "gameName":
			gameName = current.Value
		case "maxPlayers":
			maxPlayers, err = strconv.Atoi(current.Value)
			if err != nil {
				errorHandeling.PrintError(err)
				return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
			}
		default:
			return gameName, maxPlayers, fmt.Errorf("invalid number of arguments")
		}
	}

	if gameName == "" || maxPlayers == -1 {
		err := fmt.Errorf("values havent been converted")
		errorHandeling.PrintError(err)
		return gameName, maxPlayers, err
	}

	return gameName, maxPlayers, nil
}

func ConvertParamClientJoinGame(params []constants.Params, names []string) (string, error) {
	gameName := ""

	if len(params) != len(names) {
		return gameName, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return gameName, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "gameName":
			gameName = current.Value
			if !models.IsValidName(gameName) {
				err := fmt.Errorf("invalid game name")
				errorHandeling.PrintError(err)
				return gameName, err
			}
		default:
			return gameName, fmt.Errorf("invalid number of arguments")
		}
	}

	return gameName, nil
}

func ConvertParamClientSelectedCubes(params []constants.Params, names []string) ([]int, error) {
	//nested function isValidCubeValueList
	isValidCubeValueList := func(cubeValue int) bool {

		for _, scoreCube := range constants.CGScoreCubeValues {
			if cubeValue == scoreCube.Value {
				return true
			}
		}

		return false
	}

	var cubeValueList []int

	if len(params) != len(names) {
		return cubeValueList, fmt.Errorf("invalid number of arguments")
	}

	for i := 0; i < len(params); i++ {
		current := params[i]
		if current.Name != names[i] {
			return cubeValueList, fmt.Errorf("invalid number of arguments")
		}

		switch current.Name {
		case "cubeValues":
			array, err := parseParamValueArray(current.Value)
			if err != nil {
				errorHandeling.PrintError(err)
				return cubeValueList, fmt.Errorf("invalid number of arguments")
			}
			for _, value := range array {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					errorHandeling.PrintError(err)
					return cubeValueList, fmt.Errorf("invalid number of arguments")
				}

				if !isValidCubeValueList(intValue) {
					err := fmt.Errorf("not valid cube values")
					errorHandeling.PrintError(err)
					return cubeValueList, err
				}

				cubeValueList = append(cubeValueList, intValue)
			}
		default:
			return cubeValueList, fmt.Errorf("invalid number of arguments")
		}
	}

	return cubeValueList, nil
}

func parseParamValueArray(value string) ([]string, error) {
	elementName := "value"

	var valueArray []string

	if len(value) == 0 {
		return valueArray, nil
	}

	if value[0] != constants.CArrayBrackets.Opening[0] {
		return valueArray, fmt.Errorf("invalid valueArray format")
	}

	if value[len(value)-1] != constants.CArrayBrackets.Closing[0] {
		return valueArray, fmt.Errorf("invalid valueArray format")
	}

	value = value[1 : len(value)-1]

	valuesStr := strings.Split(value, constants.CParamsListElementDelimiter)

	for _, valueStr := range valuesStr {
		paramsArray, err := parseParamsStr(valueStr)
		if err != nil {
			errorHandeling.PrintError(err)
			return valueArray, fmt.Errorf("invalid valueArray format")
		}
		if len(paramsArray) != 1 {
			err := fmt.Errorf("invalid valueArray format")
			errorHandeling.PrintError(err)
			return valueArray, err
		}
		if paramsArray[0].Name != elementName {
			err := fmt.Errorf("invalid valueArray format")
			errorHandeling.PrintError(err)
			return valueArray, err
		}

		parsedValue := paramsArray[0].Value

		valueArray = append(valueArray, parsedValue)
	}

	return valueArray, nil
}

func parseParamsStr(paramsString string) ([]constants.Params, error) {
	var paramArray []constants.Params

	if len(paramsString) == 0 {
		return paramArray, nil
	}

	if paramsString[0] != constants.CParamsBrackets.Opening[0] {
		err := fmt.Errorf("invalid paramArray format")
		errorHandeling.PrintError(err)
		return paramArray, err
	}

	if paramsString[len(paramsString)-1] != constants.CParamsBrackets.Closing[0] {
		err := fmt.Errorf("invalid paramArray format")
		errorHandeling.PrintError(err)
		return paramArray, err
	}

	paramsString = paramsString[1 : len(paramsString)-1]

	paramsStr := strings.Split(paramsString, constants.CParamsDelimiter)

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

func parseParam(str string) (constants.Params, error) {
	parameter := constants.Params{}

	if len(str) == 0 {
		return parameter, nil
	}

	paramKeyValue := strings.SplitN(str, constants.CParamsKeyValueDelimiter, 2)

	parameter.Name = paramKeyValue[0]
	if parameter.Name[0] != constants.CParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}

	if parameter.Name[len(parameter.Name)-1] != constants.CParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}
	parameter.Name = parameter.Name[1 : len(parameter.Name)-1]

	parameter.Value = paramKeyValue[1]
	if parameter.Value[0] != constants.CParamsWrapper[0] {
		return parameter, fmt.Errorf("invalid param format")
	}
	if parameter.Value[len(parameter.Value)-1] != constants.CParamsWrapper[0] {
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

	opening := constants.CParamsBrackets.Opening
	closing := constants.CParamsBrackets.Closing

	playerNickname, err := extractSubstring(input, opening, closing)
	if err != nil {
		errorHandeling.PrintError(err)
		return playerNickname, playerNicknameSize, fmt.Errorf("invalid player ID format")
	}

	playerNicknameSize = len(playerNickname) + len(opening) + len(closing)

	return playerNickname, playerNicknameSize, nil
}

// endregion

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
	returnStr := constants.CParamsWrapper + name + constants.CParamsWrapper + constants.CParamsKeyValueDelimiter + constants.CParamsWrapper + value + constants.CParamsWrapper
	return returnStr
}

// endregion

// region FUNCTIONS CONVERT TO NETWORK STRING

func ConvertListGameListToNetworkString(array []*models.Game) string {
	fieldOrder := []string{"gameName", "maxPlayers", "connectedPlayers"}
	return convertListToNetworkString(array, func(element interface{}) map[string]string {
		game := element.(*models.Game)
		return map[string]string{
			"gameName":         game.GetName(),
			"maxPlayers":       fmt.Sprintf("%d", game.GetMaxPlayers()),
			"connectedPlayers": fmt.Sprintf("%d", game.GetPlayersCount()),
		}
	}, fieldOrder)
}

func ConvertListPlayerListToNetworkString(array []*models.Player) string {
	fieldOrder := []string{"playerName", "isConnected"}
	return convertListToNetworkString(array, func(element interface{}) map[string]string {
		player := element.(*models.Player)
		return map[string]string{
			"playerName":  player.GetNickname(),
			"isConnected": convertBoolToNetworkString(player.IsConnected()),
		}
	}, fieldOrder)
}

func ConvertListGameDataToNetworkString(data models.GameData) string {
	fieldOrder := []string{"playerName", "isConnected", "score", "isTurn"}
	return convertListToNetworkString(data.PlayerGameDataArr, func(element interface{}) map[string]string {
		playerGameData := element.(models.PlayerGameData)
		return map[string]string{
			"playerName":  playerGameData.Player.GetNickname(),
			"isConnected": convertBoolToNetworkString(playerGameData.Player.IsConnected()),
			"score":       fmt.Sprintf("%d", playerGameData.Score),
			"isTurn":      convertBoolToNetworkString(playerGameData.Player == data.TurnPlayer),
		}
	}, fieldOrder)
}

func ConvertListCubeValuesToNetworkString(values []int) string {
	fieldOrder := []string{"value"}
	return convertListToNetworkString(values, func(element interface{}) map[string]string {
		value := element.(int)
		return map[string]string{
			"value": fmt.Sprintf("%d", value),
		}
	}, fieldOrder)
}

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

	//End delimiter
	networkString += string(constants.CMessageEndDelimiter)

	return networkString, nil
}

func convertListToNetworkString(array interface{}, extractFields func(interface{}) map[string]string, fieldOrder []string) string {
	arrayValue := reflect.ValueOf(array)
	if arrayValue.Kind() != reflect.Slice {
		return ""
	}

	arrayStr := constants.CArrayBrackets.Opening

	for i := 0; i < arrayValue.Len(); i++ {
		element := arrayValue.Index(i).Interface()
		fields := extractFields(element)

		arrayStr += constants.CParamsBrackets.Opening
		for j, name := range fieldOrder {
			if j > 0 {
				arrayStr += constants.CParamsDelimiter
			}
			arrayStr += getArrayElementStr(name, fields[name])
		}
		arrayStr += constants.CParamsBrackets.Closing

		if i < arrayValue.Len()-1 {
			arrayStr += constants.CParamsListElementDelimiter
		}
	}

	arrayStr += constants.CArrayBrackets.Closing
	return arrayStr
}

//func ConvertListGameListToNetworkString(array []*models.Game) string {
//	nameGameName := "gameName"
//	nameMaxPlayers := "maxPlayers"
//	nameConnectedPlayers := "connectedPlayers"
//
//	arrayStr := ""
//
//	arrayStr += constants.CArrayBrackets.Opening
//
//	for i, game := range array {
//
//		valueGameName := game.GetName()
//		valueMaxPlayers := fmt.Sprintf("%d", game.GetMaxPlayers())
//		valueConnectedPlayers := fmt.Sprintf("%d", game.GetPlayersCount())
//
//		arrayStr += constants.CParamsBrackets.Opening
//		arrayStr += getArrayElementStr(nameGameName, valueGameName)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameMaxPlayers, valueMaxPlayers)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameConnectedPlayers, valueConnectedPlayers)
//		arrayStr += constants.CParamsBrackets.Closing
//
//		if i < len(array)-1 {
//			arrayStr += constants.CParamsListElementDelimiter
//		}
//	}
//
//	arrayStr += constants.CArrayBrackets.Closing
//
//	return arrayStr
//}
//
//func ConvertListPlayerListToNetworkString(array []*models.Player) string {
//	namePlayerName := "playerName"
//	nameIsConnected := "isConnected"
//
//	arrayStr := ""
//
//	arrayStr += constants.CArrayBrackets.Opening
//
//	for i, player := range array {
//
//		valuePlayerName := player.GetNickname()
//		//valueIsConnected := fmt.Sprintf("%t", player.IsConnected())
//		valueIsConnected := convertBoolToNetworkString(player.IsConnected())
//
//		arrayStr += constants.CParamsBrackets.Opening
//		arrayStr += getArrayElementStr(namePlayerName, valuePlayerName)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameIsConnected, valueIsConnected)
//		arrayStr += constants.CParamsBrackets.Closing
//
//		if i < len(array)-1 {
//			arrayStr += constants.CParamsDelimiter
//		}
//	}
//
//	arrayStr += constants.CArrayBrackets.Closing
//
//	return arrayStr
//}
//
//func ConvertListGameDataToNetworkString(data models.GameData) string {
//
//	namePlayerName := "playerName"
//	nameIsConnected := "isConnected"
//	nameScore := "score"
//	nameIsTurn := "isTurn"
//
//	arrayStr := ""
//	arrayStr += constants.CArrayBrackets.Opening
//
//	array := data.PlayerGameDataArr
//	for i, playerGameData := range array {
//		valuePlayerName := playerGameData.Player.GetNickname()
//		valueIsConnected := convertBoolToNetworkString(playerGameData.Player.IsConnected())
//		valueScore := fmt.Sprintf("%d", playerGameData.Score)
//		valueIsTurn := convertBoolToNetworkString(playerGameData.Player == data.TurnPlayer)
//
//		arrayStr += constants.CParamsBrackets.Opening
//		arrayStr += getArrayElementStr(namePlayerName, valuePlayerName)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameIsConnected, valueIsConnected)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameScore, valueScore)
//		arrayStr += constants.CParamsDelimiter
//		arrayStr += getArrayElementStr(nameIsTurn, valueIsTurn)
//		arrayStr += constants.CParamsBrackets.Closing
//
//		if i < len(array)-1 {
//			arrayStr += constants.CParamsDelimiter
//		}
//	}
//
//	arrayStr += constants.CArrayBrackets.Closing
//	return arrayStr
//}
//
//func ConvertListCubeValuesToNetworkString(values []int) string {
//	nameValue := "value"
//
//	arrayStr := ""
//
//	arrayStr += constants.CArrayBrackets.Opening
//
//	for i, value := range values {
//
//		valueValue := fmt.Sprintf("%d", value)
//
//		arrayStr += constants.CParamsBrackets.Opening
//		arrayStr += getArrayElementStr(nameValue, valueValue)
//		arrayStr += constants.CParamsBrackets.Closing
//		if i < len(values)-1 {
//			arrayStr += constants.CParamsDelimiter
//		}
//	}
//
//	arrayStr += constants.CArrayBrackets.Closing
//
//	return arrayStr
//}

func convertBoolToNetworkString(value bool) string {
	//true to 1, false to 0
	if value {
		return "1"
	}
	return "0"
}

func convertPlayerNicknameToNetworkString(nickname string) string {
	networkStr := ""
	networkStr += constants.CParamsBrackets.Opening
	networkStr += nickname
	networkStr += constants.CParamsBrackets.Closing

	return networkStr
}

func convertParamsToNetworkString(params []constants.Params) (string, error) {
	length := len(params)
	if length == 0 {
		return constants.CParamsBrackets.Opening + constants.CParamsBrackets.Closing, nil
	}

	paramsStr := constants.CParamsBrackets.Opening
	for i, param := range params {
		paramStr := _convertParamElementToNetworkString(param.Name, param.Value)
		paramsStr += paramStr
		if i < length-1 {
			paramsStr += constants.CParamsDelimiter
		}
	}

	paramsStr += constants.CParamsBrackets.Closing

	return paramsStr, nil
}

func _convertParamElementToNetworkString(name string, value string) string {
	return constants.CParamsWrapper + name + constants.CParamsWrapper + constants.CParamsKeyValueDelimiter + constants.CParamsWrapper + value + constants.CParamsWrapper
}

// endregion

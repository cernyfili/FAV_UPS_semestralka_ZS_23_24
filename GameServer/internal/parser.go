package internal

import (
	"errors"
	"regexp"
	"strconv"
)

var gRegexPatterns = map[string]string{
	"parseGameIDPlayerID":  `^\{"gameID":"(\d+)","playerID":"(\d+)"\}$`,
	"parsePlayerLoginArgs": `^\{"nickname":"([A-Za-z0-9_\\-]+)"\}$`,
	"parseCreateGameArgs":  `^\{"name":"([A-Za-z0-9_\\-]+)","maxPlayers":"(\d+)"\}$`,
	"parsePlayerID":        `^\{"playerID":"(\d+)"\}$`,
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
		return gameID, playerID, errors.New("invalid args format")
	}

	gameID, err := strconv.Atoi(matches[1])
	if err != nil {
		return gameID, playerID, errors.New("invalid maxPlayers format")
	}

	playerID, err = strconv.Atoi(matches[2])
	if err != nil {
		return gameID, playerID, errors.New("invalid maxPlayers format")
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
		return playerID, errors.New("invalid args format")
	}

	playerID, err := strconv.Atoi(matches[1])
	if err != nil {
		return playerID, errors.New("invalid playerID format")
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
		return nickname, errors.New("invalid args format")
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
		return name, maxPlayers, errors.New("invalid args format")
	}

	name = matches[1]
	maxPlayers, err := strconv.Atoi(matches[2])
	if err != nil {
		return name, maxPlayers, errors.New("invalid maxPlayers format")
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

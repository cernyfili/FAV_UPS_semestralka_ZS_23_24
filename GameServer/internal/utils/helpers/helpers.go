package helpers

import (
	"encoding/json"
	"fmt"
	"gameserver/internal/models"
	"gameserver/internal/utils/errorHandeling"
	"io"
	"net"
	"os"
)

func RemovePlayerFromLists(player *models.Player) error {

	playerlist := models.GetInstancePlayerList()

	gamelist := models.GetInstanceGameList()
	//Remove player from playerlist
	err := playerlist.RemoveItem(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create playersGame %w", err)
	}

	//Remove player from playersGame

	gameFromList := gamelist.GetPlayersGame(player)

	err = gameFromList.RemovePlayer(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("cannot create playersGame %w", err)
	}

	return nil
}

func RemovePlayerFromGame(player *models.Player) error {
	err := RemovePlayerFromLists(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("Error sending response: %w", err)
	}

	return nil
}

func RemovePlayerFromList(list []*models.Player, player *models.Player) []*models.Player {
	var newList []*models.Player
	for _, p := range list {
		if p != player {
			newList = append(newList, p)
		}
	}
	return newList
}

type ServerConfig struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// read config file in json format where is server - which has values ip_adress and port
func ReadConfigFile(filePath string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("could not open config file: %w", err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("could not read config file: %w", err)
	}

	var config struct {
		Server ServerConfig `json:"server"`
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return "", "", fmt.Errorf("could not unmarshal config file: %w", err)
	}

	//close file
	err = file.Close()
	if err != nil {
		return "", "", fmt.Errorf("could not close config file: %w", err)
	}

	//check if ip and port are valid
	if !isValidIP(config.Server.IP) {
		return "", "", fmt.Errorf("invalid ip address")
	}

	if !isValidPort(config.Server.Port) {
		return "", "", fmt.Errorf("invalid port")
	}

	portStr := fmt.Sprintf("%d", config.Server.Port)

	return config.Server.IP, portStr, nil
}

func isValidPort(port int) bool {
	return port > 0 && port <= 65535
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

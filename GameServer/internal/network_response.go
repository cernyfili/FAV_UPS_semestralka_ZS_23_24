package internal

import "gameserver/internal/utils"

func SendErrorToPlayer(i int, err error) {
	//todo
}

func SendPlayerConnectResponse(playerID int, gameList utils.GameList) error {
	//todo
	return nil
}

func SendCreateGameResponse(id int, game utils.Game) error {
	//todo
	return nil
}

func SendJoinGameResponse(id int, game utils.Game) error {
	//todo
	return nil
}

func SendStartGameResponse(id int, game utils.Game) error {
	//todo
	return nil
}

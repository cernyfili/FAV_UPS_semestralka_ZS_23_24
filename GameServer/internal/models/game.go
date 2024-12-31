package models

import (
	"fmt"
	"gameserver/internal/utils"
	"math/rand"
	"sync"
)

// region CONSTANTS
const (
	cMinimumPlayers = 2
	cMaxCubeCount   = 6

	cCubeMinValue = 1
	cCubeMaxValue = 6
)

//endregion

//region DATA STRUCTURES

// enum for game state
type GameState int

const (
	Running GameState = iota
	Created
	Ended
)

type Throw struct {
	cubeValues           []int
	selectedCubesIndexes []int
}

type Turn struct {
	ThrowArr []Throw
}

// Player Data Structure represents plyaer data for game
type PlayerGameData struct {
	Player      *Player
	Score       int
	TurnHistory []Turn
}

// Game represents a game with a unique ID, a list of g_players, and turn data.
type Game struct {
	gameID             int
	name               string
	playersGameDataArr []PlayerGameData
	maxPlayers         int
	turnNum            int
	gameStateValue     GameState
	mutex              sync.Mutex
}

type GameData struct {
	PlayerGameDataArr []PlayerGameData
	TurnPlayer        *Player
}

//endregion

//region FUNCTIONS

// CreateGame creates a new game with a unique ID and initializes Player and turn slices.
func CreateGame(name string, maxPlayers int) (*Game, error) {
	//Check if the arguments are valid
	if name == "" || maxPlayers <= 1 {
		return nil, fmt.Errorf("invalid arguments")
	}

	return &Game{
		name:               name,
		playersGameDataArr: make([]PlayerGameData, 0),
		maxPlayers:         maxPlayers,
		turnNum:            0,
		gameStateValue:     Created,
	}, nil
}

func generateCubeValues(count int) []int {
	array := make([]int, count)
	for i := 0; i < count; i++ {
		array[i] = rand.Intn(cCubeMaxValue) + cCubeMinValue
	}
	return array
}

// NextTurn returns the next Player in turn.
func (g *Game) NextTurn() (*Player, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if len(g.playersGameDataArr) == 0 {
		return nil, fmt.Errorf("no players in the game")
	}
	if g.gameStateValue != Running {
		return nil, fmt.Errorf("game is not running")
	}

	currentPlayerIndex := g.turnNum % len(g.playersGameDataArr)
	nextPlayerIndex := (currentPlayerIndex + 1) % len(g.playersGameDataArr)

	g.turnNum++

	return g.playersGameDataArr[nextPlayerIndex].Player, nil
}

//region PLAYER

// AddPlayer adds a new Player to the game.
func (g *Game) AddPlayer(pPlayer *Player) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if len(g.playersGameDataArr) >= g.maxPlayers {
		return fmt.Errorf("game is full")
	}

	playerGameData := PlayerGameData{
		Player:      pPlayer,
		Score:       0,
		TurnHistory: make([]Turn, 0),
	}

	g.playersGameDataArr = append(g.playersGameDataArr, playerGameData)
	pPlayer.game = g

	return nil
}

// getTurnPlayer returns the current turn player
func (g *Game) getTurnPlayer() (*Player, error) {

	if len(g.playersGameDataArr) == 0 {
		return nil, fmt.Errorf("no players in the game")
	}
	if g.gameStateValue != Running {
		return nil, fmt.Errorf("game is not running")
	}

	currentPlayerIndex := g.turnNum % len(g.playersGameDataArr)

	return g.playersGameDataArr[currentPlayerIndex].Player, nil
}

func (g *Game) HasPlayer(player *Player) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for _, p := range g.playersGameDataArr {
		if p.Player == player {
			return true
		}
	}
	return false
}

func (g *Game) RemovePlayer(player *Player) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for i, p := range g.playersGameDataArr {
		if p.Player == player {
			player.game = nil
			g.playersGameDataArr = append(g.playersGameDataArr[:i], g.playersGameDataArr[i+1:]...)
			break
		}
	}
	return nil
}

//endregion

//region GETTERS

// Get name
func (g *Game) GetName() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.name
}

// Get max players
func (g *Game) GetMaxPlayers() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.maxPlayers
}

// Get players count
func (g *Game) GetPlayersCount() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return len(g.playersGameDataArr)
}

// Get gameID
func (g *Game) GetGameID() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.gameID
}

func (g *Game) GetState() GameState {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.gameStateValue
}

// GetPlayers returns the players in the game
func (g *Game) GetPlayers() []*Player {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var players []*Player
	for _, p := range g.playersGameDataArr {
		players = append(players, p.Player)
	}
	return players
}

func (g *Game) GetGameData() (GameData, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var gameData GameData
	gameData.PlayerGameDataArr = g.playersGameDataArr
	turnPlayer, err := g.NextTurn()
	if err != nil {
		return GameData{}, err
	}

	gameData.TurnPlayer = turnPlayer

	return gameData, nil
}

func (g *Game) GetScoreIncrease(cubeIndexes []int, player *Player) (int, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	cubeValues, err := g.getLastThrowCubeValues(player)
	if err != nil {
		return 0, fmt.Errorf("failed to get last throw")
	}

	if len(cubeIndexes) == 0 {
		return 0, fmt.Errorf("no cubes selected")
	}

	if len(cubeIndexes) > len(cubeValues) {
		return 0, fmt.Errorf("invalid cube indexes")
	}

	scoreIncrease := 0
	for _, index := range cubeIndexes {
		if index < 0 || index >= len(cubeValues) {
			return 0, fmt.Errorf("invalid cube indexes")
		}
		value := cubeValues[index]
		isScoreValue := false
		for _, scoreCube := range utils.CGScoreCubeValues {
			if value == scoreCube.Value {
				scoreIncrease += scoreCube.ScoreValue
				isScoreValue = true
				break
			}
		}

		if !isScoreValue {
			return 0, fmt.Errorf("invalid cube value")
		}
	}

	return scoreIncrease, nil
}

func (g *Game) GetPlayerScore(player *Player) (int, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	playerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		return 0, fmt.Errorf("player not found")
	}

	return playerGameData.Score, nil
}

func (g *Game) IsFull() bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return len(g.playersGameDataArr) >= g.maxPlayers
}

func (g *Game) IsEnoughPlayers() bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return len(g.playersGameDataArr) >= cMinimumPlayers
}

func (g *Game) getPlayerGameData(player *Player) (PlayerGameData, error) {

	for _, gameData := range g.playersGameDataArr {
		if gameData.Player == player {
			return gameData, nil
		}
	}

	return PlayerGameData{}, fmt.Errorf("player not found")
}

// GetLastThrow returns the last throw of the player
func (g *Game) getLastThrowCubeValues(player *Player) ([]int, error) {

	playerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		return nil, fmt.Errorf("player not found")
	}

	turnHistory := playerGameData.TurnHistory
	if len(turnHistory) == 0 {
		return nil, fmt.Errorf("no throws")
	}

	turn := turnHistory[len(turnHistory)-1]
	throwArr := turn.ThrowArr
	if len(throwArr) == 0 {
		return nil, fmt.Errorf("no throws")
	}

	throw := throwArr[len(throwArr)-1]
	return throw.cubeValues, nil
}

//endregion

// region SETTERS
func (g *Game) SetState(state GameState) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.gameStateValue = state
}

func (g *Game) NewThrow(player *Player) ([]int, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	turnPlayer, err := g.getTurnPlayer()
	if err != nil {
		return nil, fmt.Errorf("not your turn")
	}
	if player != turnPlayer {
		return nil, fmt.Errorf("not your turn")
	}

	turnPlayerGameData, err := g.getPlayerGameData(turnPlayer)
	if err != nil {
		return nil, fmt.Errorf("player not found")
	}

	var cubeCount int
	if len(turnPlayerGameData.TurnHistory) == 0 {
		cubeCount = cMaxCubeCount
	} else {
		lastTurn := turnPlayerGameData.TurnHistory[len(turnPlayerGameData.TurnHistory)-1]
		lastThrow := lastTurn.ThrowArr[len(lastTurn.ThrowArr)-1]

		cubeCount = cMaxCubeCount - len(lastThrow.selectedCubesIndexes)
	}

	cubeValues, err := g.addThrow(turnPlayer, cubeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to add throw")
	}

	return cubeValues, nil
}

func (g *Game) addThrow(player *Player, count int) ([]int, error) {

	turnPlayerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		return nil, fmt.Errorf("player not found")
	}

	var throw Throw

	throw.cubeValues = generateCubeValues(count)
	throw.selectedCubesIndexes = make([]int, 0)

	//get the last turn
	turnHistory := turnPlayerGameData.TurnHistory
	if len(turnHistory) == 0 {
		turnHistory = make([]Turn, 0)
	}
	turn := turnHistory[len(turnHistory)-1]
	throwArr := turn.ThrowArr
	if len(throwArr) == 0 {
		throwArr = make([]Throw, 0)
	}

	//set the new throw
	throwArr = append(throwArr, throw)
	turn.ThrowArr = throwArr
	turnHistory[len(turnHistory)-1] = turn
	turnPlayerGameData.TurnHistory = turnHistory

	return throw.cubeValues, nil
}

func (g *Game) SetPlayerScore(player *Player, score int) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	playerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		return fmt.Errorf("player not found")
	}

	playerGameData.Score = score

	return nil
}

//endregion

//endregion

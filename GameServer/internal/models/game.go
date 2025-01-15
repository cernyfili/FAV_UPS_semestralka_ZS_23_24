package models

import (
	"fmt"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
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

// nextTurn returns the next Player in turn.
func (g *Game) nextTurn() (*Player, error) {

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

	return nil
}

// GetTurnPlayer returns the current turn player
func (g *Game) GetTurnPlayer() (*Player, error) {
	//lock
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.getTurnPlayer()
}

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

// IsPlayerTurn returns true if it is the player's turn
func (g *Game) IsPlayerTurn(player *Player) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	turnPlayer, err := g.getTurnPlayer()
	if err != nil {
		return false
	}

	return turnPlayer == player
}

// ShiftTurn shifts the turn to the next player
func (g *Game) ShiftTurn() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.shiftTurn()
}

func (g *Game) shiftTurn() error {

	_, err := g.nextTurn()
	if err != nil {
		return fmt.Errorf("failed to shift turn")
	}

	return nil
}

func (g *Game) GetGameData() (GameData, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var gameData GameData
	gameData.PlayerGameDataArr = g.playersGameDataArr
	turnPlayer, err := g.getTurnPlayer()
	if err != nil {
		errorHandeling.PrintError(err)
		return GameData{}, err
	}

	gameData.TurnPlayer = turnPlayer

	return gameData, nil
}

func (g *Game) GetScoreIncrease(cubeValuesList []int, player *Player) (int, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	playerLastThrowCubeValues, err := g.getLastThrowCubeValues(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return 0, fmt.Errorf("failed to get last throw")
	}

	// check if cubeValuesList elements are all in playerLastThrowCubeValues
	for _, cubeValue := range cubeValuesList {
		isInList := false
		for _, playerCubeValue := range playerLastThrowCubeValues {
			if cubeValue == playerCubeValue {
				isInList = true
				break
			}
		}
		if !isInList {
			return 0, fmt.Errorf("invalid cube value")
		}
	}

	scoreIncrease := 0
	for _, cubeValue := range cubeValuesList {
		isScoreValue := false
		for _, scoreCube := range constants.CGScoreCubeValues {
			if cubeValue == scoreCube.Value {
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
		errorHandeling.PrintError(err)
		return 0, fmt.Errorf("player not found")
	}

	return playerGameData.Score, nil
}

func (g *Game) IsFull() bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return len(g.playersGameDataArr) >= g.maxPlayers
}

func (g *Game) isEnoughPlayers() bool {
	return len(g.playersGameDataArr) >= cMinimumPlayers
}

func (g *Game) getPlayerGameData(player *Player) (*PlayerGameData, error) {

	for _, gameData := range g.playersGameDataArr {
		if gameData.Player == player {
			return &gameData, nil
		}
	}

	return &PlayerGameData{}, fmt.Errorf("player not found")
}

// GetLastThrow returns the last throw of the player
func (g *Game) getLastThrowCubeValues(player *Player) ([]int, error) {

	playerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		errorHandeling.PrintError(err)
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

// start Game
func (g *Game) StartGame() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.gameStateValue != Created {
		return fmt.Errorf("game has already started or ended")
	}

	if !g.isEnoughPlayers() {
		return fmt.Errorf("not enough players to start the game")
	}

	g.gameStateValue = Running
	g.turnNum = 0

	err := g.shiftTurn()
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) NewThrow(player *Player) ([]int, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	turnPlayer, err := g.getTurnPlayer()
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, fmt.Errorf("not your turn")
	}
	if player != turnPlayer {
		return nil, fmt.Errorf("not your turn")
	}

	turnPlayerGameData, err := g.getPlayerGameData(turnPlayer)
	if err != nil {
		errorHandeling.PrintError(err)
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
		errorHandeling.PrintError(err)
		return nil, fmt.Errorf("failed to add throw")
	}

	return cubeValues, nil
}

func (g *Game) addThrow(player *Player, count int) ([]int, error) {

	turnPlayerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return nil, fmt.Errorf("player not found")
	}

	var throw Throw

	throw.cubeValues = generateCubeValues(count)
	throw.selectedCubesIndexes = make([]int, 0)

	turnNum := g.turnNum

	// Add new turn if it is the first throw of the turn
	if len(turnPlayerGameData.TurnHistory) < turnNum {
		emptyTurn := Turn{}
		emptyTurn.ThrowArr = make([]Throw, 0)
		turnPlayerGameData.TurnHistory = append(turnPlayerGameData.TurnHistory, emptyTurn)
	}
	if len(turnPlayerGameData.TurnHistory) != turnNum {
		return nil, fmt.Errorf("failed to add throw")
	}

	array := turnPlayerGameData.TurnHistory[turnNum-1].ThrowArr
	array = append(array, throw)
	turnPlayerGameData.TurnHistory[turnNum-1].ThrowArr = array

	// Set the updated turn history in the game playersGameDataArr
	for i, p := range g.playersGameDataArr {
		if p.Player == player {
			g.playersGameDataArr[i].TurnHistory = turnPlayerGameData.TurnHistory
			break
		}
	}

	return throw.cubeValues, nil
}

func (g *Game) SetPlayerScore(player *Player, score int) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	playerGameData, err := g.getPlayerGameData(player)
	if err != nil {
		errorHandeling.PrintError(err)
		return fmt.Errorf("player not found")
	}

	playerGameData.Score = score

	return nil
}

//endregion

//endregion

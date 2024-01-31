- Start
- Waiting for request
- Get request

## Control player table

- player list

  - player

    - player_id
    - game_id

## Requests

```json
{
  "header": {
    "signature": "KIVUPS",
    "command_id": 4,
    "player_id" : 5
  },
  "params": [
    {
      "key": "nickname",
      "value": "super_karel"
    }
  ]
}
```

### Message format

```json
KIVUPS;command_id;player_id;args...
```

- size_bytes: 1 bytes
- KIVUPS: 6 bytes
- command_id: 1 byte
- player_id: 2 bytes
- args: based on size_bytes

### Client-Server

#### **1 - Player connect**

- _Format:_
  - **{"nickname":"super_karel"}**

#### **21 - RESPONSE**

- one player
- _Format:_
  - **{"player_id":"3"}**

#### **2 - Create game**

- createGame()
- _Format:_

  - **ADD_GAME;{"name":"Room1","max_player_count":"5"}**
- ANSWER ALL - player joined to game

  - game id
  - player list

    - player_name
    - player_score

#### **30 SUCCESS - RESPONSE**

- one player

#### **3 - Join game**

- _Format:_
  - **JOIN_GAME;{"game_id":"7"}**
  -
- joinGame(game_id)
- ANSWER ALL - player joined to game

#### **23 RESPONSE**

- all players in game
- _Format:_
  - **["karel_1", "karel_2", "karel_3", "karel_4", "karel_5"]**
  -

#### **4 - Start game**

- - _Format:_
- **START_GAME;{}**
- Game.startGame()
- ANSWER ALL - game started
-

#### **5 - Player rolling dice**

- - - _Format:_
  - **ROLL_DICE;{}**
  - ANSWER ALL - game update

    - player_turn

      - score
      - current_dice

#### **7 - Player logging out**

- - _Format:_
- **LEAVE_GAME;{}**
- ANSWER ALL - player logged out

  - player_id
- player ending game

  - ANSWER ALL - player ended game

    - game_id
- Game

  - game state

    - in room
    - running
  - player list

    - player

      - current score
      - number of turns
  - game turn number
  -

#### **6 - End game**

- - _Format:_
- **LEAVE_GAME;{}**

#### **8 - Get game score**

- - _Format:_
- **LEAVE_GAME;{"game_id":"6"}**

  ### Server-Client
- **Rooms list - response: PLAYER_CONNECT**

  - game in room list
  - _Format:_
    - **ROOM_LIST;["game_id":{"name":"Room1","player_count":"5","state":"WAITING"}]}**

---

# Player Finite Automata

```mermaid
stateDiagram
  Start --> Lobby: ClientLogin
  
  Lobby --> End: ClientLogout
  Lobby --> Lobby: ServerUpdateGameList
  Lobby --> Game: ClientJoinGame
  Lobby --> Game: ClientCreateGame
  Lobby --> Error_lobby: ErrorPlayerUnreachable

  Error_lobby --> Lobby: ServerReconnectGameList
  
  Game --> Running_Game: ClientStartGame
  Game --> Running_Game: ServerUpdateStartGame
  Game --> Game: ServerUpdatePlayerList
  Game --> Error_game: ErrorPlayerUnreachable
  
  Error_game --> Game: ServerReconnectPlayerList
  Error_game --> Error_running_Game: ServerUpdateStartGame

  Running_Game --> Lobby: ServerUpdateEndScore
  Running_Game --> My_turn: ServerStartTurn
  Running_Game --> Running_Game: ServerUpdateGameData
  Running_Game --> Error_running_Game: ErrorPlayerUnreachable
  
  Error_running_Game --> Running_Game: ServerReconnectGameData
  Error_running_Game --> Error_game: ServerUpdateEndScore

  My_turn --> Fork_my_turn: ClientRollDice
  My_turn --> My_turn: ServerPingPlayer
  My_turn --> Error_running_Game: ErrorPlayerUnreachable
  
  Fork_my_turn --> Running_Game: ResponseServerDiceEndTurn
  Fork_my_turn --> Next_dice: ResponseServerDiceNext

  Next_dice --> Running_Game: ClientEndTurn
  Next_dice --> Fork_next_dice: ClientNextDice
  Next_dice --> Next_dice: ServerPingPlayer
  
  Fork_next_dice --> My_turn: ResponseServerNextDiceSuccess
  Fork_next_dice --> Lobby: ResponseServerNextDiceEndScore
```


```mermaid
stateDiagram
    Start --> Lobby : CLIENT->SERVER ClientLogin (nickname)\n -- SERVER-CLIENT ResponseServerGameList\n -or- SERVER->CLIENT ResponseServerError \n -or- SERVER->CLIENT ResponseServerErrDuplicitNickname
    Lobby --> End : CLIENT->SERVER ClientLogout \n -- SERVER->CLIENT ResponseServerSuccess \n -or- SERVER->CLIENT ResponseServerError
    Lobby --> Lobby : SERVER->CLIENT ServerUpdateGameList (game_list (game_name, max_players, connected_players_num)) \n CLIENT->SERVER ResponseServerSuccess
    Lobby --> Game : CLIENT->SERVER ClientJoinGame (game_id) \n -- SERVER->CLIENT ResponseServerSuccess  \n -or- SERVER->CLIENT ResponseServerError \n SERVER->PLAYERS_IN_GAME ServerUpdatePlayerList (player_list (active/inactive))
    Lobby --> Game : CLIENT->SERVER ClientCreateGame (game_name, max_players) \n -- SERVER->CLIENT ResponseServerSuccess \n -or- SERVER->CLIENT ResponseServerError  \n SERVER->PLAYERS_IN_LOBBY ServerUpdateGameList(gameList (game_name, max_players, connected_players_num))
    Lobby --> Error_lobby : ErrorPlayerUnreachable
    
    Error_lobby --> Lobby : SERVER->CLIENT ServerReconnectGameList (game_list (game_name, max_players, connected_players_num)) \n CLIENT->SERVER success
    
    Game --> Running_Game : CLIENT->SERVER ClientStartGame \n -- SERVER->CLIENT  ResponseServerSuccess \n -or- SERVER->CLIENT ResponseServerError\n SERVER->PLAYERS_IN_GAME ServerUpdateStartGame(player_list (active/inactive), turn_player, score)
    Game --> Running_Game : Somebody_started_game\n \n SERVER->CLIENT ServerUpdateStartGame \n CLIENT->SERVER ResponseServerSuccess
    Game --> Game : SERVER->CLIENT ServerUpdatePlayerList (player_list (active/inactive)) \n CLIENT->SERVER ResponseServerSuccess
    Game --> Error_game : ErrorPlayerUnreachable
    
    Error_game --> Game : SERVER->CLIENT ServerReconnectPlayerList (player_list (active/inactive)) \n CLIENT->SERVER ResponseServerSuccess
    Error_game --> Error_running_Game : SERVER->CLIENT ServerUpdateStartGame \n CLIENT->SERVER ResponseServerSuccess
    
    Running_Game --> Lobby : SERVER->CLIENT ServerUpdateEndScore (winner_player_nickname, game_list) \n CLIENT->SERVER ResponseServerSuccess
    Running_Game --> My_turn : SERVER->CLIENT ServerStartTurn \n CLIENT->SERVER ResponseServerSuccess
    Running_Game --> Running_Game : SERVER->CLIENT ServerUpdateGameData(player_list (active/inactive), turn_player, score) \n CLIENT->SERVER ResponseServerSuccess
    Running_Game --> Error_running_Game : ErrorPlayerUnreachable
    
    Error_running_Game --> Running_Game : SERVER->CLIENT ServerReconnectGameData(player_list (active/inactive), turn_player, score) \n CLIENT->SERVER ResponseServerSuccess
    Error_running_Game --> Error_game : SERVER->CLIENT ServerUpdateEndScore (winner_player_nickname, game_list) \n CLIENT->SERVER ResponseServerSuccess
    
    My_turn --> fork_my_turn : CLIENT->SERVER ClientRollDice
    My_turn --> My_turn : SERVER->CLIENT ServerPingPlayer \n CLIENT->SERVER ResponseServerSuccess
    My_turn --> Error_running_Game : ErrorPlayerUnreachable
    
    fork_my_turn --> Running_Game : SERVER->CLIENT ResponseServerDiceEndTurn(cubes_values, score) \n SERVER->PLAYERS_IN_GAME ServerGameUpdates(player_list (active/inactive), turn_player, score)
    fork_my_turn --> Next_dice : SERVER->CLIENT ResponseServerDiceNext(dice_result- cubes_values)
    
    
    Next_dice --> Running_Game : CLIENT->SERVER ClientEndTurn() \n SERVER->CLIENT endTurn(score) \n SERVER->PLAYERS_IN_GAME ServerGameUpdates(player_list (active/inactive), turn_player, score)
    Next_dice --> Fork_next_dice : CLIENT->SERVER ClientNextDice(selected_cubes)
    Next_dice --> Next_dice : SERVER->CLIENT ServerPingPlayer \n CLIENT->SERVER ResponseServerSuccess
    
    Fork_next_dice --> My_turn : SERVER->CLIENT ResponseServerNextDiceSuccess()
    Fork_next_dice --> Lobby : SERVER->CLIENT ResponseServerNextDiceEndScore() \n SERVER->PLAYERS_IN_GAME ServerUpdateEndScore (winner_player_nickname, game_list)
    
```

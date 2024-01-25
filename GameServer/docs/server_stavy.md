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
  Error_lobby --> Lobby: ServerReconnectGameList
  Lobby --> Error_lobby: ErrorPlayerUnreachable

  Game --> Running_Game: ClientStartGame
  Game --> Running_Game: ServerPlayerStartGame
  Game --> Game: ServerUpdatePlayerList
  Game --> Error_game: ErrorPlayerUnreachable
  Error_game --> Game: ServerReconnectPlayerList
  Error_game --> Error_running_Game: ServerPlayerStartGame

  Running_Game --> Lobby: ServerPlayerEndScore
  Running_Game --> My_turn: ServerStartTurn
  Running_Game --> Running_Game: ServerUpdateGameData
  Running_Game --> Error_running_Game: ErrorPlayerUnreachable
  Error_running_Game --> Running_Game: ServerReconnectGameData
  Error_running_Game --> Error_game: ServerPlayerEndScore


  My_turn --> Fork_my_turn: ClientRollDice
  My_turn --> My_turn: ServerPingPlayer
  My_turn --> Error_running_Game: ErrorPlayerUnreachable
  Fork_my_turn --> Lobby: ResponseServerDiceEndScore
  Fork_my_turn --> Running_Game: ResponseServerDiceEndTurn
  Fork_my_turn --> Next_dice: ResponseServerDiceNext

  Next_dice --> Running_Game: ResponseClientEndTurn
  Next_dice --> My_turn: ResponseClientNextDice
```


```mermaid
stateDiagram
    [*] --> Lobby : CLIENT->SERVER player_login (nickname)\n SERVER-CLIENT games_list --\n CLIENT-SERVER success
    Lobby --> [*] : CLIENT->SERVER player_logout \n SERVER->CLIENT success
    Lobby --> Lobby : SERVER->CLIENT ServerUpdateGameList (game_list (game_name, max_players, connected_players_num)) \n CLIENT->SERVER success
    Lobby --> Game : CLIENT->SERVER join_game (game_id) \n SERVER->CLIENT success  \n ERROR player_disconnect --\n SERVER->PLAYERS_IN_GAME ServerUpdatePlayerList (player_list (active/inactive))
    Lobby --> Game : CLIENT->SERVER create_game (game_name, max_players) \n SERVER->CLIENT success - -  \n SERVER->PLAYERS_IN_LOBBY ServerUpdateGameList(gameList (game_name, max_players, connected_players_num))
    Game --> Running_Game : CLIENT->SERVER start_game \n SERVER->CLIENT success --\n SERVER->PLAYERS_IN_GAME game_started(player_list (active/inactive), turn_player, score)
    Game --> Running_Game : Somebody_started_game\n \n SERVER->CLIENT game_started \n CLIENT->SERVER success
    Game --> Game : SERVER->CLIENT ServerUpdatePlayerList (player_list (active/inactive)) \n CLIENT->SERVER success
    Running_Game --> Lobby : SERVER->CLIENT end_score_reached (winner_player_nickname, game_list) \n CLIENT->SERVER success
    Running_Game --> My_turn : SERVER->CLIENT start_turn \n CLIENT->SERVER success \n ServerGameUpdates(player_list (active/inactive), turn_player, score)
    Running_Game --> Running_Game : SERVER->CLIENT ServerGameUpdates(player_list (active/inactive), turn_player, score) \n CLIENT->SERVER success
    
    
    My_turn --> fork_my_turn : CLIENT->SERVER roll_dice
    fork_my_turn --> Lobby : SERVER->CLIENT end_score(winner_player_nickname, game_list) \n CLIENT->SERVER success \n  SERVER->PLAYERS_IN_GAME end_score_reached (winner_player_nickname, game_list)
    fork_my_turn --> Running_Game : SERVER->CLIENT dice_end_turn(cubes_values, score) \n CLIENT->SERVER success \n SERVER->PLAYERS_IN_GAME ServerGameUpdates(player_list (active/inactive), turn_player, score)
    fork_my_turn --> Next_dice : SERVER->CLIENT next_dice(dice_result- cubes_values) \n CLIENT->SERVER success 
    
    
    Next_dice --> Running_Game : CLIENT->SERVER end_turn() \n SERVER->CLIENT endTurn(score) \n SERVER->PLAYERS_IN_GAME ServerGameUpdates(player_list (active/inactive), turn_player, score)
    Next_dice --> My_turn : CLIENT->SERVER next_dice(selected_cubes) \n SERVER->CLIENt success
```

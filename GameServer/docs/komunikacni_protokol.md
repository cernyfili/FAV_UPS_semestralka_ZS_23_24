## Client

- game_create
  
  - game_name
  
  - player_id

- game_join
  
  - game_id
  
  - player_id

- game_start
  
  - game_id
  
  - player_id

- game_leave
  
  - game_id
  
  - player_id

- dice_roll
  
  - player_id
  
  - turn_number

- turn_end
  
  - player_id
  
  - turn_number

## Server

- game_add_player
  
  - player_id
  
  - game_id

- game_start
  
  - game_id

- turn_start
  
  - turn_number
  
  - player_id

- dice_result
  
  - list n{cube1_value, ...., cubeN_value}

- game_end
  
  - game_id

- game_update
  
  - player_id
  
  - dice_result

# Headers

- game

- dice

- turn
# Command Handlers
from constants import CommandTypeEnum

command_handlers = {
    CommandTypeEnum.ClientCreateGame.value.id: {"handler": process_client_create_game, "command": CommandTypeEnum.ClientCreateGame.value},
    CommandTypeEnum.ClientJoinGame.value.id: {"handler": process_client_join_game, "command": CommandTypeEnum.ClientJoinGame.value},
    CommandTypeEnum.ClientStartGame.value.id: {"handler": process_client_start_game, "command": CommandTypeEnum.ClientStartGame.value},
    CommandTypeEnum.ClientLogout.value.id: {"handler": process_client_player_logout, "command": CommandTypeEnum.ClientLogout.value},
    CommandTypeEnum.ClientReconnect.value.id: {"handler": process_client_reconnect, "command": CommandTypeEnum.ClientReconnect.value},
}

def process_message(message, conn):
    # Check if valid signature
    if message.signature != "CMessageSignature":
        raise ValueError("Invalid signature")

    command_id = message.command_id
    player_nickname = message.player_nickname
    timestamp = message.timestamp
    params = message.parameters

    connection_info = {
        "connection": conn,
        "timestamp": timestamp,
    }

    # If player_login
    if command_id == CommandTypeEnum.ClientLogin.value.id:
        err = process_player_login(player_nickname, connection_info, CommandTypeEnum.ClientLogin.value, params)
        if err:
            raise ValueError(f"Error sending response: {err}")

    # Check if values valid
    command_info = command_handlers.get(command_id)
    if command_info is None:
        raise ValueError("Invalid command or incorrect number of arguments")

    # Get player
    player = g_player_list.get_item(player_nickname)
    if player is None:
        raise ValueError("Invalid command or incorrect number of arguments")

    # Set connection to player
    player.set_connection_info(connection_info)

    # Call the corresponding handler function
    err = command_info["handler"](player, params, command_info["command"])
    if err:
        raise ValueError("Invalid command or incorrect number of arguments")

    return None


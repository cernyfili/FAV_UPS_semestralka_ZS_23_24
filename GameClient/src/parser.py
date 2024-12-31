from typing import List

from GUI import Game

TIMESTAMP_SIZE = 32

COMMANDID_SIZE = 2

SIGNATURE_SIZE = 6


class Brackets:
    def __init__(self, opening, closing):
        self.opening = opening
        self.closing = closing


c_params_brackets = Brackets("{", "}")
c_array_brackets = Brackets("[", "]")
c_params_delimiter = ","
c_params_key_value_delimiter = ":"
c_params_wrapper = "\""



class Param:
    def __init__(self, name: str, value: str):
        self.name = name
        self.value = value

class NetworkMessage:
    def __init__(self, signature: str, command_id: int, timestamp: str, player_nickname: str, parameters: List[Param]):
        self.signature = signature
        self.command_id = command_id
        self.timestamp = timestamp
        self.player_nickname = player_nickname
        self.parameters = parameters



def convert_params_to_network_string(params: List[Param]) -> str:
    params_str = c_params_brackets.opening
    for param in params:
        params_str += c_params_wrapper + param.name + c_params_wrapper + c_params_key_value_delimiter + c_params_wrapper + param.value + c_params_wrapper + c_params_delimiter
    params_str = params_str.rstrip(c_params_delimiter)  # remove the trailing delimiter
    params_str += c_params_brackets.closing

    return params_str

def parse_message(input_str):
    signature_size = SIGNATURE_SIZE
    command_id_size = COMMANDID_SIZE
    timestamp_size = TIMESTAMP_SIZE
    min_size_message = signature_size + command_id_size + timestamp_size

    if len(input_str) < min_size_message:
        raise ValueError("Invalid input format")

    start = 0

    # Read Signature
    signature = input_str[start : start + signature_size]
    start += signature_size
    start += 1

    # Read Command ID
    command_id = input_str[start : start + command_id_size]
    start += command_id_size
    start += 1

    # Read Timestamp
    timestamp = input_str[start : start + timestamp_size]
    start += timestamp_size
    start += 1

    # Read nickname
    player_nickname, player_nickname_size = parse_message_player_id(input_str[start:])
    start += player_nickname_size
    start += 1

    # Read Parameters
    parameters_str = input_str[start:]

    params = parse_params_str(parameters_str)

    # Convert values
    command_id_int = int(command_id)

    return NetworkMessage(signature, command_id_int, timestamp, player_nickname, params)


def parse_params_str(params_string):
    param_array = []

    if len(params_string) == 0:
        return param_array

    if params_string[0] != c_params_brackets.opening:
        raise ValueError("Invalid paramArray format")

    if params_string[-1] != c_params_brackets.closing:
        raise ValueError("Invalid paramArray format")

    params_string = params_string[1:-1]

    params_str = params_string.split(c_params_delimiter)

    for param_str in params_str:
        parameter = parse_param(param_str)
        param_array.append(parameter)

    return param_array


def parse_param(str):
    if len(str) == 0:
        return Param("", "")

    param_key_value = str.split(c_params_key_value_delimiter)

    if len(param_key_value) != 2:
        raise ValueError("Invalid param format")

    name = param_key_value[0]
    if name[0] != c_params_wrapper or name[-1] != c_params_wrapper:
        raise ValueError("Invalid param format")
    name = name[1:-1]

    value = param_key_value[1]
    if value[0] != c_params_wrapper or value[-1] != c_params_wrapper:
        raise ValueError("Invalid param format")
    value = value[1:-1]

    return Param(name, value)


def parse_message_player_id(input):
    player_nickname = ""
    player_nickname_size = 0

    if len(input) == 0:
        raise ValueError("Invalid player ID format")

    opening = c_params_brackets.opening
    closing = c_params_brackets.closing

    player_nickname = input[len(opening):-len(closing)]
    player_nickname_size = len(player_nickname) + len(opening) + len(closing)

    return player_nickname, player_nickname_size

def convert_message_to_network_string(message: NetworkMessage) -> str:
    network_string = ""

    signature_size = SIGNATURE_SIZE
    command_id_size = COMMANDID_SIZE
    timestamp_size = TIMESTAMP_SIZE

    # Signature
    if len(message.signature) != signature_size:
        raise ValueError("Invalid signature")
    network_string += message.signature

    # Command ID
    command_id_str = str(message.command_id)
    if len(command_id_str) != command_id_size:
        raise ValueError("Invalid command ID")
    network_string += command_id_str

    # Timestamp
    if len(message.timestamp) != timestamp_size:
        raise ValueError("Invalid timestamp")
    network_string += message.timestamp

    # PlayerNickname
    network_string += c_params_brackets.opening
    network_string += message.player_nickname
    network_string += c_params_brackets.closing

    # Parameters
    params_str = convert_params_to_network_string(message.parameters)
    network_string += params_str

    return network_string


def parse_game(game_str):
    game = Game("", 0)
    if len(game_str) == 0:
        return game

    if game_str[0] != c_params_brackets.opening:
        raise ValueError("Invalid game format")

    if game_str[-1] != c_params_brackets.closing:
        raise ValueError("Invalid game format")

    game_str = game_str[1:-1]

    game_str = game_str.split(c_params_key_value_delimiter)

    if len(game_str) != 2:
        raise ValueError("Invalid game format")

    name = game_str[0]
    if name[0] != c_params_wrapper or name[-1] != c_params_wrapper:
        raise ValueError("Invalid game format")
    name = name[1:-1]

    max_players = game_str[1]
    if max_players[0] != c_params_wrapper or max_players[-1] != c_params_wrapper:
        raise ValueError("Invalid game format")
    max_players = max_players[1:-1]

    player_count game_str[2]
    if player_count[0] != c_params_wrapper or player_count[-1] != c_params_wrapper:
        raise ValueError("Invalid game format")

    game = Game(name, max_players, player_count)
    #todo check if wrong

    return game


def convert_params_game_list(parameters: List[Param]) -> List[Game]:
    # Convert the parameters to a list of Game objects
    games = []
    if len(parameters) != 0:
        return games
    param_value = parameters[0].value

    if param_value[0] != c_array_brackets.opening or param_value[-1] != c_array_brackets.closing:
        raise ValueError("Invalid param value format")
    param_value = param_value[1:-1]

    param_value = param_value.split(c_params_delimiter)
    for game_str in param_value:

        game = parse_game(game_str)
        games.append(game)

    return games
from typing import List

from shared.constants import CMessagePartsSizes, CMessageConfig, CCommandTypeEnum
from shared.data_structures import Game, Param, NetworkMessage, GameList, PlayerList, GameData, Player, PlayerGameData


def parse_message(input_str) -> NetworkMessage:
    def _parse_params_str(params_string):
        def process_array_values(param_array):
            for param in param_array:
                if param.value[0] == CMessageConfig.ARRAY_BRACKETS.opening and param.value[-1] == CMessageConfig.ARRAY_BRACKETS.closing:
                    array_str = param.value[1:-1]
                    array = array_str.split(CMessageConfig.PARAMS_DELIMITER)

                    for element in array:
                        param_array.append(_parse_param_element(element))

            return param_array
        def _parse_param_element(str):
            if len(str) == 0:
                return Param("", "")

            param_key_value = str.split(CMessageConfig.PARAMS_KEY_VALUE_DELIMITER)

            if len(param_key_value) != 2:
                raise ValueError("Invalid param format")

            name = param_key_value[0]
            if name[0] != CMessageConfig.PARAMS_WRAPPER or name[-1] != CMessageConfig.PARAMS_WRAPPER:
                raise ValueError("Invalid param format")
            name = name[1:-1]

            value = param_key_value[1]
            if value[0] != CMessageConfig.PARAMS_WRAPPER or value[-1] != CMessageConfig.PARAMS_WRAPPER:
                raise ValueError("Invalid param format")
            value = value[1:-1]

            return Param(name, value)

        param_array = []

        if len(params_string) == 0:
            return param_array

        if params_string[0] != CMessageConfig.PARAMS_BRACKETS.opening:
            raise ValueError("Invalid paramArray format")

        if params_string[-1] != CMessageConfig.PARAMS_BRACKETS.closing:
            raise ValueError("Invalid paramArray format")

        params_string = params_string[1:-1]

        params_str = params_string.split(CMessageConfig.PARAMS_DELIMITER)

        for param_str in params_str:
            parameter = _parse_param_element(param_str)
            param_array.append(parameter)

        param_array = process_array_values(param_array)

        return param_array

    def convert_arrayparam_to_specified_datastructures(command_id: int, params: List) -> List[Param]:
        command = CCommandTypeEnum.get_command_by_id(command_id)
        if command.MessageParamListInfo is None:
            return params

        if len(params) != 1:
            return params

        # always only one list
        param_array = params[0]

        # check names in array elements
        for param_element_array in param_array:
            elements_names = [param.name for param in param_element_array]
            if elements_names != command.MessageParamListInfo.param_names:
                raise ValueError("Invalid parameter names in array element")

        converted_data = command.MessageParamListInfo.convert_function(param_array)

        return_list = []
        return_list.append(Param(param_array.name, converted_data))

        return return_list


    def _parse_message_player_id(input):
        player_nickname = ""
        player_nickname_size = 0

        if len(input) == 0:
            raise ValueError("Invalid player ID format")

        opening = CMessageConfig.PARAMS_BRACKETS.opening
        closing = CMessageConfig.PARAMS_BRACKETS.closing

        player_nickname = input[len(opening):-len(closing)]
        player_nickname_size = len(player_nickname) + len(opening) + len(closing)

        return player_nickname, player_nickname_size

    signature_size = CMessagePartsSizes.SIGNATURE_SIZE
    command_id_size = CMessagePartsSizes.COMMANDID_SIZE
    timestamp_size = CMessagePartsSizes.TIMESTAMP_SIZE
    min_size_message = signature_size + command_id_size + timestamp_size

    if len(input_str) < min_size_message:
        raise ValueError("Invalid input format")

    start = 0

    # Read Signature
    signature = input_str[start : start + signature_size]
    if signature != CMessageConfig.SIGNATURE:
        raise ValueError("Invalid signature")
    start += signature_size
    start += 1

    # Read Command ID
    command_id = input_str[start : start + command_id_size]
    if not command_id.isdigit() or int(command_id) < 0 or CCommandTypeEnum.is_command_id_in_enum(int(command_id)):
        raise ValueError("Invalid command ID")
    start += command_id_size
    start += 1

    # Read Timestamp
    timestamp = input_str[start : start + timestamp_size]
    start += timestamp_size
    start += 1

    # Read nickname
    player_nickname, player_nickname_size = _parse_message_player_id(input_str[start:])
    start += player_nickname_size
    start += 1

    # Read Parameters
    parameters_str = input_str[start:]

    params = _parse_params_str(parameters_str)

    # Convert values
    command_id_int = int(command_id)

    params = convert_arrayparam_to_specified_datastructures(command_id_int, params)

    return NetworkMessage(signature, command_id_int, timestamp, player_nickname, params)

def convert_message_to_network_string(message: NetworkMessage) -> str:
    def _convert_params_to_network_string(params: List[Param]) -> str:
        params_str = CMessageConfig.PARAMS_BRACKETS.opening
        for param in params:
            params_str += CMessageConfig.PARAMS_WRAPPER + param.name + CMessageConfig.PARAMS_WRAPPER + CMessageConfig.PARAMS_KEY_VALUE_DELIMITER + CMessageConfig.PARAMS_WRAPPER + param.value + CMessageConfig.PARAMS_WRAPPER + CMessageConfig.PARAMS_DELIMITER
        params_str = params_str.rstrip(CMessageConfig.PARAMS_DELIMITER)  # remove the trailing delimiter
        params_str += CMessageConfig.PARAMS_BRACKETS.closing

        return params_str

    network_string = ""

    signature_size = CMessagePartsSizes.SIGNATURE_SIZE
    command_id_size = CMessagePartsSizes.COMMANDID_SIZE
    timestamp_size = CMessagePartsSizes.TIMESTAMP_SIZE

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
    network_string += CMessageConfig.PARAMS_BRACKETS.opening
    network_string += message.player_nickname
    network_string += CMessageConfig.PARAMS_BRACKETS.closing

    # Parameters
    params_str = _convert_params_to_network_string(message.parameters)
    network_string += params_str

    return network_string

#
# def parse_game(game_str):
#     game = Game("", 0)
#     if len(game_str) == 0:
#         return game
#
#     if game_str[0] != c_params_brackets.opening:
#         raise ValueError("Invalid game format")
#
#     if game_str[-1] != c_params_brackets.closing:
#         raise ValueError("Invalid game format")
#
#     game_str = game_str[1:-1]
#
#     game_str = game_str.split(c_params_key_value_delimiter)
#
#     if len(game_str) != 2:
#         raise ValueError("Invalid game format")
#
#     name = game_str[0]
#     if name[0] != c_params_wrapper or name[-1] != c_params_wrapper:
#         raise ValueError("Invalid game format")
#     name = name[1:-1]
#
#     max_players = game_str[1]
#     if max_players[0] != c_params_wrapper or max_players[-1] != c_params_wrapper:
#         raise ValueError("Invalid game format")
#     max_players = max_players[1:-1]
#
#     player_count game_str[2]
#     if player_count[0] != c_params_wrapper or player_count[-1] != c_params_wrapper:
#         raise ValueError("Invalid game format")
#
#     game = Game(name, max_players, player_count)
#     #todo check if wrong
#
#     return game


# def convert_params_game_list(parameters: List[Param]) -> List[Game]:
#     # Convert the parameters to a list of Game objects
#     games = []
#     if len(parameters) != 0:
#         return games
#     param_value = parameters[0].value
#
#     if param_value[0] != CMessageConfig.ARRAY_BRACKETS.opening or param_value[-1] != CMessageConfig.ARRAY_BRACKETS.closing:
#         raise ValueError("Invalid param value format")
#     param_value = param_value[1:-1]
#
#     param_value = param_value.split(CMessageConfig.PARAMS_DELIMITER)
#     for game_str in param_value:
#         game = parse_game(game_str)
#         games.append(game)
#
#     return games

class Convertor:
    def convert_param_list_to_game_list(self, param_array : List[List[Param]]) -> GameList:
        game_list = GameList([])
        for element_array in param_array:
            # get element value with name "name"
            name = [param.value for param in element_array if param.name == "gameName"][0]
            connected_count_players = [param.value for param in element_array if param.name == "connectedPlayers"][0]
            max_players = [param.value for param in element_array if param.name == "maxPlayers"][0]

            game = Game(name, connected_count_players, max_players)
            game_list.append(game)

        return game_list

    def convert_param_list_to_player_list(self, param_array : List[List[Param]]) -> PlayerList:
        player_list = PlayerList([])
        for element_array in param_array:
            # get element value with name "name"
            name = [param.value for param in element_array if param.name == "playerName"][0]

            player = Player(name)
            player_list.append(player)

        return player_list

    def convert_param_list_to_game_data(self, param_array : List[List[Param]]) -> GameData:
        game_data = GameData([])
        for element_array in param_array:
            # get element value with name "player_name"
            player_name = [param.value for param in element_array if param.name == "playerName"][0]
            is_connected = [param.value for param in element_array if param.name == "isConnected"][0]
            score = [param.value for param in element_array if param.name == "score"][0]
            is_turn = [param.value for param in element_array if param.name == "isTurn"][0]

            player_game_data = PlayerGameData(player_name, is_connected, score, is_turn)
            game_data.append(player_game_data)

        return game_data
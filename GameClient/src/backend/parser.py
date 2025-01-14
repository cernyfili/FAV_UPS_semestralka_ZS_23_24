from typing import List

from shared.constants import CMessagePartsSizes, CMessageConfig, CCommandTypeEnum, Param, NetworkMessage


def parse_message(input_str: str) -> NetworkMessage:
    def _parse_param_str(params_string)->Param:
        def __process_array_values(param : Param) -> Param:
            def ___is_array_param(param : Param) -> bool:
                if len(param.value) == 0:
                    return False
                return param.value[0] == CMessageConfig.ARRAY_BRACKETS.opening and param.value[-1] == CMessageConfig.ARRAY_BRACKETS.closing
            def ___parse_array_element(array_element: str)-> List[Param]:

                param_list = []
                # {"gameName": "Game1", "maxPlayers": "4", "connectedPlayers": "2"}
                if array_element[0] != CMessageConfig.PARAMS_BRACKETS.opening or array_element[-1] != CMessageConfig.PARAMS_BRACKETS.closing:
                    raise ValueError("Invalid array element format")
                array_element = array_element[1:-1]
                # "gameName": "Game1", "maxPlayers": "4", "connectedPlayers": "2"
                values_array = array_element.split(CMessageConfig.PARAMS_DELIMITER)
                for value in values_array:
                    param_list.append(__parse_param_element(value))

                return param_list
            # if not array
            if not ___is_array_param(param):
                return param

            array_str = param.value[1:-1]
            if len(array_str) == 0:
                return Param(param.name, [])

            array = array_str.split(CMessageConfig.PARAMS_ARRAY_DELIMITER)

            value_array = []
            for element in array:
                value_array.append(___parse_array_element(element))

            return Param(param.name,value_array)
        def __parse_param_element(str):
            if len(str) == 0:
                return Param("", "")

            param_key_value = str.split(CMessageConfig.PARAMS_KEY_VALUE_DELIMITER, 1)

            name = param_key_value[0]
            if name[0] != CMessageConfig.PARAMS_WRAPPER or name[-1] != CMessageConfig.PARAMS_WRAPPER:
                raise ValueError("Invalid param format")
            name = name[1:-1]

            value = param_key_value[1]
            if value[0] != CMessageConfig.PARAMS_WRAPPER or value[-1] != CMessageConfig.PARAMS_WRAPPER:
                raise ValueError("Invalid param format")
            value = value[1:-1]

            return Param(name, value)

        if len(params_string) == 0:
            return Param("", "")

        if params_string[0] != CMessageConfig.PARAMS_BRACKETS.opening:
            raise ValueError("Invalid paramArray format")

        if params_string[-1] != CMessageConfig.PARAMS_BRACKETS.closing:
            raise ValueError("Invalid paramArray format")

        # without brackets
        params_string = params_string[1:-1] #always one parameter

        parameter = __parse_param_element(params_string)

        param = __process_array_values(parameter)

        return param

    def _convert_arrayparam_to_specified_datastructures(command_id: int, param: Param) -> List[Param]:
        command = CCommandTypeEnum.get_command_by_id(command_id)
        message_param_list_info = command.param_list_info
        if message_param_list_info is None:
            return [param]

        param_value = param.value

        if len(param_value) == 0:
            return [param]
        if not isinstance(param_value, list):
            raise [param]

        param_array = param_value
        # check names in array elements
        for element in param_array:
            elements_names = [param.name for param in element]
            if elements_names != message_param_list_info.param_names:
                raise ValueError("Invalid parameter names in array element")
        converted_data = message_param_list_info.convert_function(param_array)
        # todo raise error which then closes connection becouse message is not in format that is expected

        return [Param(param.name, converted_data)]


    def _parse_message_player_nickname(input):
        player_nickname = ""
        player_nickname_size = 0

        if len(input) == 0:
            raise ValueError("Invalid player ID format")

        opening = CMessageConfig.PARAMS_BRACKETS.opening
        closing = CMessageConfig.PARAMS_BRACKETS.closing

        player_nickname = input[len(opening):input.find(closing)]
        player_nickname_size = len(player_nickname) + len(opening) + len(closing)

        return player_nickname, player_nickname_size

    if len(input_str) == 0:
        raise ValueError("Invalid input format")

    signature_size = CMessagePartsSizes.SIGNATURE_SIZE
    command_id_size = CMessagePartsSizes.COMMANDID_SIZE
    timestamp_size = CMessagePartsSizes.TIMESTAMP_SIZE
    min_size_message = signature_size + command_id_size + timestamp_size

    if len(input_str) < min_size_message:
        raise ValueError("Invalid input format")

    start = 0

    # Check if the message ends with the end of message character
    if input_str[-1] != CMessageConfig.END_OF_MESSAGE:
        raise ValueError("Invalid message format")

    # Check if not more end of message characters
    if input_str[:-1].find(CMessageConfig.END_OF_MESSAGE) != -1:
        raise ValueError("Buffer read multiple messages")

    # remove the end of message character
    input_str = input_str[:-1]

    # Read Signature
    signature = input_str[start : start + signature_size]
    if signature != CMessageConfig.SIGNATURE:
        raise ValueError("Invalid signature")
    start += signature_size

    # Read Command ID
    command_id = input_str[start : start + command_id_size]
    if not command_id.isdigit() or int(command_id) < 0 or not CCommandTypeEnum.is_command_id_in_enum(int(command_id)):
        raise ValueError("Invalid command ID")
    start += command_id_size

    # Read Timestamp
    timestamp = input_str[start : start + timestamp_size]
    start += timestamp_size

    # Read nickname
    player_nickname, player_nickname_size = _parse_message_player_nickname(input_str[start:])
    start += player_nickname_size

    # Read Parameters
    parameters_str = input_str[start:]

    param = _parse_param_str(parameters_str)

    # Convert values
    command_id_int = int(command_id)

    param = _convert_arrayparam_to_specified_datastructures(command_id_int, param)

    return NetworkMessage(signature, command_id_int, timestamp, player_nickname, param)

def convert_message_to_network_string(message: NetworkMessage) -> str:
    def _convert_params_to_network_string(params: List[Param]) -> str:
        params_str = CMessageConfig.PARAMS_BRACKETS.opening
        for param in params:
            param_name = str(param.name)
            param_value = str(param.value)
            params_str += CMessageConfig.PARAMS_WRAPPER + param_name + CMessageConfig.PARAMS_WRAPPER + CMessageConfig.PARAMS_KEY_VALUE_DELIMITER + CMessageConfig.PARAMS_WRAPPER + param_value + CMessageConfig.PARAMS_WRAPPER + CMessageConfig.PARAMS_DELIMITER
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
    command_id_str = str(format(message.command_id, '02d'))
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

    # Add end of message
    network_string += CMessageConfig.END_OF_MESSAGE

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

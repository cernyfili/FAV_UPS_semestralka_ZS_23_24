from dataclasses import dataclass
from enum import Enum
from typing import Final, List

from src.backend.state_manager import GameStateMachine

# region Utils constants

LOGS_FOLDER_PATH: Final = "logs"


# endregion

# region APP LOGIC

# region DATA STRUCTURES

@dataclass
class MessageParamListInfo:
    param_names: list[str]
    convert_function: callable

@dataclass
class ScoreCube:
    Value: int
    ScoreValue: int

@dataclass
class Param:
    name: str
    value: any

@dataclass
class Command:
    id: int
    trigger: callable
    param_names: List[str]
    param_list_info: any

    def is_valid_param_names(self, params: List[Param]) -> bool:
        if not self.param_names:
            return True
        if len(params) != len(self.param_names):
            return False
        param_names = [param.name for param in params]
        return self.param_names == param_names

@dataclass(frozen=True)
class Brackets:
    opening: str
    closing: str



@dataclass
class Game:
    name: str
    connected_count_players: int
    max_players: int



class GameList(list):
    _ELEMENT_DATA_TYPE = Game

    def __init__(self, elements):
        if not all(isinstance(element, self._ELEMENT_DATA_TYPE) for element in elements):
            raise ValueError("All elements must be integers")
        super().__init__(elements)


@dataclass
class Player:
    name: str
    is_connected: bool


class PlayerList(list):
    _ELEMENT_DATA_TYPE = Player

    def __init__(self, elements):
        if not all(isinstance(element, self._ELEMENT_DATA_TYPE) for element in elements):
            raise ValueError("All elements must be integers")
        super().__init__(elements)


@dataclass
class PlayerGameData:
    player_name: str
    is_connected: bool
    score: int
    is_turn: bool


class GameData(list):
    _ELEMENT_DATA_TYPE = PlayerGameData

    def __init__(self, elements):
        if not all(isinstance(element, self._ELEMENT_DATA_TYPE) for element in elements):
            raise ValueError("All elements must be integers")
        super().__init__(elements)

class CubeValuesList(list):
    _ELEMENT_DATA_TYPE = int

    def __init__(self, elements):
        if not all(isinstance(element, self._ELEMENT_DATA_TYPE) for element in elements):
            raise ValueError("All elements must be integers")
        super().__init__(elements)




@dataclass
class NetworkMessage:

    def __init__(self, signature: str, command_id: int, timestamp, player_nickname: str, parameters: List[
        Param]):
        def process_signature(signature: str) -> str:
            if signature != CMessageConfig.SIGNATURE:
                raise ValueError("Invalid signature")
            return signature

        def process_command_id(command_id: int) -> int:
            if not CCommandTypeEnum.is_command_id_in_enum(command_id):
                raise ValueError("Invalid command ID")
            return command_id

        def process_player_nickname(player_nickname: str) -> str:
            if not CMessageConfig.is_valid_name(player_nickname):
                raise ValueError("Invalid player nickname")
            return player_nickname

        self._signature : str = process_signature(signature)
        self._command_id : int = process_command_id(command_id)
        self._timestamp : timestamp = timestamp
        self._player_nickname : str = process_player_nickname(player_nickname)
        self._parameters : List[Param] = self._process_parameters(parameters)

    def _process_parameters(self, parameters: List[Param]) -> List[Param]:
        command = CCommandTypeEnum.get_command_by_id(self._command_id)
        if not command.is_valid_param_names(parameters):
            raise ValueError("Invalid parameters")
        return parameters

    @property
    def signature(self) -> str:
        return self._signature

    @property
    def command_id(self) -> int:
        return self._command_id

    @property
    def timestamp(self):
        return self._timestamp

    @property
    def player_nickname(self) -> str:
        return self._player_nickname

    @property
    def parameters(self) -> List[Param]:
        return self._parameters

    def get_param_value_by_name(self, param_name: str):
        for param in self._parameters:
            if param.name == param_name:
                return param.value
        raise ValueError("Parameter not found")

    def get_array_param(self):
        if len(self._parameters) != 1:
            raise ValueError("Invalid number of parameters")

        param_value = self._parameters[0].value

        if param_value == '[]':
            return []

        if not isinstance(param_value, list):
            raise ValueError("Invalid parameter type")

        return param_value

    def get_single_param(self):
        if len(self._parameters) != 1:
            raise ValueError("Invalid number of parameters")

        return self._parameters[0].value

    def __str__(self):
        command_name = CCommandTypeEnum.get_command_name_from_id(self._command_id)
        return f"NetworkMessage: {self._signature}, {self._command_id}:{command_name}, {self._timestamp}, {self._player_nickname}, {self._parameters}"



# endregion

# region CONSTANTS
GAME_STATE_MACHINE : GameStateMachine = GameStateMachine()

def reset_game_state_machine():
    global GAME_STATE_MACHINE
    GAME_STATE_MACHINE = GameStateMachine()

NETWORK_PARAM_EMPTY : Final = []


class Combination(list):
    _ELEMENT_DATA_TYPE = int

    def __init__(self, elements):
        if not all(isinstance(element, self._ELEMENT_DATA_TYPE) for element in elements):
            raise ValueError("All elements must be integers")
        super().__init__(elements)


class CombinationList:

    def __init__(self, combinations: List[Combination]):
        self.list = combinations

    @staticmethod
    def is_combination_in_list(combination: Combination, cube_list) -> bool:
        cube_list_copy = cube_list.copy()
        for value in combination:
            if value in cube_list_copy:
                cube_list_copy.remove(value)
            else:
                return False
        return True

    def create_allowed_values_mask(self, cube_values: CubeValuesList) -> List[bool]:
        def _set_true_for_combination(mask: List[bool], combination: Combination):
            for value in combination:
                for i, v in enumerate(cube_values):
                    if v == value:
                        mask[i] = True

            return mask


        mask = [False] * len(cube_values)
        for combination in self.list:
            # if cube_values have all values from combination
            if self.is_combination_in_list(combination, cube_values):
               mask = _set_true_for_combination(mask, combination)

        return mask

    def __str__(self):
        return_str = ""
        for combination in self.list:
            return_str += f"{combination}\n"

        return return_str

ALLOWED_CUBE_VALUES_COMBINATIONS : CombinationList = CombinationList([
    Combination([1]),
    Combination([5]),
    # Combination([2, 2])
])

SCORE_VALUES_CUBES : Final = [
    ScoreCube(1, 100),
    ScoreCube(5, 50),
]

@dataclass(frozen=True)
class CGameConfig:
    MIN_PLAYERS : Final = 2
    MAX_PLAYERS : Final = 4
    MAX_ROUNDS : Final = 10
    MAX_DICE_ROLLS : Final = 3


class Convertor:
    @staticmethod
    def _validate_name(name: str) -> bool:
        return CMessageConfig.is_valid_name(name)

    @staticmethod
    def convert_param_list_to_game_list(param_array : List[List[Param]]) -> GameList:

        def validate_connected_players(connected_players: int) -> bool:
            """
            :param connected_players: int - The number of players currently connected to the game.
            :return: bool - True if the number of connected players is valid, False otherwise.
            """
            if not connected_players:
                return False
            if connected_players < 0:
                return False
            if connected_players > CGameConfig.MAX_PLAYERS:
                return False
            return True

        def validate_max_players(max_players: int) -> bool:
            if not max_players:
                return False
            if max_players < CGameConfig.MIN_PLAYERS:
                return False
            if max_players > CGameConfig.MAX_PLAYERS:
                return False
            return True
        
        game_list = GameList([])
        for element_array in param_array:
            # get element value with name "name"
            try:
                name = [param.value for param in element_array if param.name == "gameName"][0]
                connected_count_players = int([param.value for param in element_array if param.name == "connectedPlayers"][0])
                max_players = int([param.value for param in element_array if param.name == "maxPlayers"][0])
            except Exception as e:
                raise ValueError("Invalid game list format")
            
            if not Convertor._validate_name(name) or not validate_connected_players(connected_count_players) or not validate_max_players(max_players):
                raise ValueError("Invalid game list format")

            game = Game(name, connected_count_players, max_players)
            game_list.append(game)

        return game_list

    @staticmethod
    def convert_param_list_to_player_list(param_array : List[List[Param]]) -> PlayerList:
        player_list = PlayerList([])
        for element_array in param_array:
            # get element value with name "name"
            name = [param.value for param in element_array if param.name == "playerName"][0]
            is_connected = bool(int([param.value for param in element_array if param.name == "isConnected"][0]))
            if not Convertor._validate_name(name):
                raise ValueError("Invalid player list format")
            
            player = Player(name, is_connected)
            player_list.append(player)

        return player_list

    @staticmethod
    def convert_param_list_to_game_data(param_array : List[List[Param]]) -> GameData:
        def validate_score(score: int) -> bool:
            if score < 0:
                return False
            return True

        game_data = GameData([])
        for element_array in param_array:
            # get element value with name "player_name"
            player_name = [param.value for param in element_array if param.name == "playerName"][0]
            is_connected = bool(int([param.value for param in element_array if param.name == "isConnected"][0]))
            score = int([param.value for param in element_array if param.name == "score"][0])
            is_turn = bool(int([param.value for param in element_array if param.name == "isTurn"][0]))

            if not Convertor._validate_name(player_name) or not validate_score(score):
                raise ValueError("Invalid game data format")

            player_game_data = PlayerGameData(player_name, is_connected, score, is_turn)
            game_data.append(player_game_data)

        return game_data

    @staticmethod
    def convert_cube_values_to_list(param_array : List[List[Param]]) -> CubeValuesList:
        def validate_cube_value(value: int) -> bool:
            if value < 1 or value > 6:
                return False
            return True

        cube_values = CubeValuesList([])
        for element_array in param_array:
            # get element value with name "value"
            value = int([param.value for param in element_array if param.name == "value"][0])
            if not validate_cube_value(value):
                raise ValueError("Invalid cube value")

            cube_values.append(value)

        return cube_values

game_list_info = MessageParamListInfo(["gameName", "maxPlayers", "connectedPlayers"], Convertor.convert_param_list_to_game_list)
player_list_info = MessageParamListInfo(["playerName","isConnected"], Convertor.convert_param_list_to_player_list)
game_data_info = MessageParamListInfo(["playerName", "isConnected", "score", "isTurn"], Convertor.convert_param_list_to_game_data)
cube_values_info = MessageParamListInfo(["value"], Convertor.convert_cube_values_to_list)

class CCommandTypeEnum(Enum):
    # CLIENT->SERVER
    ClientLogin: Command = Command(1, GAME_STATE_MACHINE.ClientLogin, [], None)
    ClientCreateGame: Command = Command(2, GAME_STATE_MACHINE.ClientCreateGame, ["gameName", "maxPlayers"], None)
    ClientJoinGame: Command = Command(3, GAME_STATE_MACHINE.ClientJoinGame, ["gameName"], None)
    ClientStartGame: Command = Command(4, GAME_STATE_MACHINE.ClientStartGame, [], None)
    ClientRollDice: Command = Command(5, GAME_STATE_MACHINE.ClientRollDice, [], None)
    ClientLogout: Command = Command(7, None, [], None)

    ClientReconnect: Command = Command(8, GAME_STATE_MACHINE.ClientReconnect, [], None)

    ClientSelectedCubes: Command = Command(61, GAME_STATE_MACHINE.ClientNextDice, ["cubeValues"], cube_values_info)
    ClientEndTurn: Command = Command(62, GAME_STATE_MACHINE.ClientEndTurn, [], None)

    # RESPONSES SERVER->CLIENT
    ResponseServerSuccess: Command = Command(30, None, [], None)
    ResponseServerError: Command = Command(32, None, ["message"], None)

    ResponseServerGameList: Command = Command(33, GAME_STATE_MACHINE.ResponseServerGameList, ["gameList"],
                                              game_list_info)

    ResponseServerSelectCubes: Command = Command(34, GAME_STATE_MACHINE.ResponseServerDiceNext, ["cubeValues"], cube_values_info)
    ResponseServerEndTurn: Command = Command(35, GAME_STATE_MACHINE.ResponseServerDiceEndTurn, [], None)

    ResponseServerEndScore: Command = Command(36, GAME_STATE_MACHINE.ResponseServerNextDiceEndScore, [], None)
    ResponseServerDiceSuccess: Command = Command(37, GAME_STATE_MACHINE.ResponseServerNextDiceSuccess, [], None)

    ## Reconnect
    ResponseServerReconnectBeforeGame: Command = Command(46, GAME_STATE_MACHINE.ResponseServerReconnectBeforeGame,
                                                         ["gameList"],
                                                         game_list_info)
    ResponseServerReconnectRunningGame: Command = Command(47, GAME_STATE_MACHINE.ResponseServerReconnectRunningGame,
                                                          ["gameData"],
                                                          game_data_info)

    # SERVER->CLIENT
    ## SERVER -> ALL CLIENTS
    ServerUpdateStartGame: Command = Command(41, GAME_STATE_MACHINE.ServerUpdateStartGame, [], None)
    ServerUpdateEndScore: Command = Command(42, GAME_STATE_MACHINE.ServerUpdateEndScore, ["playerName"], None)
    ServerUpdateNotEnoughPlayers: Command = Command(51, GAME_STATE_MACHINE.ServerUpdateNotEnoughPlayers, [], None)

    ServerUpdateGameData: Command = Command(43, GAME_STATE_MACHINE.ServerUpdateGameData, ["gameData"], game_data_info)
    ServerUpdateGameList: Command = Command(44, GAME_STATE_MACHINE.ServerUpdateGameList, ["gameList"], game_list_info)
    ServerUpdatePlayerList: Command = Command(45, GAME_STATE_MACHINE.ServerUpdatePlayerList, ["playerList"], player_list_info)

    ## SERVER -> SINGLE CLIENT
    ServerStartTurn: Command = Command(49, GAME_STATE_MACHINE.ServerStartTurn, [], None)
    ServerPingPlayer: Command = Command(50, GAME_STATE_MACHINE.ServerPingPlayer, [], None)

    # RESPONSES CLIENT->SERVER
    ResponseClientSuccess: Command = Command(60, None, [], None)

    @staticmethod
    def is_command_id_in_enum(command_id: int) -> bool:
        for command in CCommandTypeEnum:
            if command.value.id == command_id:
                return True
        return False

    @classmethod
    def get_command_by_id(cls, _command_id) -> Command:
        if not cls.is_command_id_in_enum(_command_id):
            raise ValueError("Invalid command ID")

        for element in cls:
            if element.value.id == _command_id:
                return element.value
    @staticmethod
    def get_command_name_from_id(command_id):
        for command in CCommandTypeEnum:
            if command.value.id == command_id:
                return command.name
        return None


@dataclass(frozen=True)
class CMessagePartsSizes:
    SIGNATURE_SIZE: Final = 6
    COMMANDID_SIZE: Final = 2
    TIMESTAMP_SIZE: Final = 26


@dataclass(frozen=True)
class CMessageConfig:
    PARAMS_BRACKETS: Brackets = Brackets("{", "}")
    ARRAY_BRACKETS: Brackets = Brackets("[", "]")
    PARAMS_DELIMITER: str = ","
    PARAMS_ARRAY_DELIMITER: str = ";"
    PARAMS_KEY_VALUE_DELIMITER: str = ":"
    PARAMS_WRAPPER: str = "\""
    TIMESTAMP_FORMAT: str = '%Y-%m-%d %H:%M:%S.%f'
    SIGNATURE: str = "KIVUPS"
    END_OF_MESSAGE: str = "\n"
    NAME_MIN_CHARS : Final = 3
    NAME_MAX_CHARS : Final = 20

    @staticmethod
    def is_valid_name(name: str) -> bool:
        if not name:
            return False

        if not name.isalnum():
            return False

        return CMessageConfig.NAME_MIN_CHARS <= len(name) <= CMessageConfig.NAME_MAX_CHARS

@dataclass(frozen=True)
class CNetworkConfig:
    RECEIVE_TIMEOUT = 5
    BUFFER_SIZE: Final = 1024
    MAX_MESSAGE_SIZE: Final = 1024
    RECONNECT_ATTEMPTS: Final = 3  # todo change
    RECONNECT_TIMEOUT_SEC: Final = 2

# endregion
# endregion

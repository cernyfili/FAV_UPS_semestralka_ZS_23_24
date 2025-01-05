import threading
from dataclasses import dataclass
from enum import Enum
from typing import Final, List, TypeAlias

import timestamp


from backend.state_manager import GameStateMachine

# region DATA STRUCTURES

@dataclass
class MessageParamListInfo:
    param_names: list[str]
    convert_function: callable = None

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


GameList: TypeAlias = list[Game]


@dataclass
class Player:
    name: str


PlayerList: TypeAlias = list[Player]


@dataclass
class PlayerGameData:
    player_name: str
    is_connected: bool
    score: int
    is_turn: bool


GameData: TypeAlias = list[PlayerGameData]


@dataclass
class NetworkMessage:

    def __init__(self, signature: str, command_id: int, timestamp: timestamp, player_nickname: str, parameters: List[
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
    def timestamp(self) -> timestamp:
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

    def __str__(self):
        command_name = CCommandTypeEnum.get_command_name_from_id(self._command_id)
        return f"NetworkMessage: {self._signature}, {self._command_id}:{command_name}, {self._timestamp}, {self._player_nickname}, {self._parameters}"



# endregion

GAME_STATE_MACHINE : GameStateMachine = GameStateMachine()

def reset_game_state_machine():
    global GAME_STATE_MACHINE
    GAME_STATE_MACHINE = GameStateMachine()

NETWORK_PARAM_EMPTY : Final = []

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
    def convert_param_list_to_game_list(param_array : List[List[Param]]) -> GameList:
        game_list = GameList([])
        for element_array in param_array:
            # get element value with name "name"
            name = [param.value for param in element_array if param.name == "gameName"][0]
            connected_count_players = [param.value for param in element_array if param.name == "connectedPlayers"][0]
            max_players = [param.value for param in element_array if param.name == "maxPlayers"][0]

            game = Game(name, connected_count_players, max_players)
            game_list.append(game)

        return game_list

    @staticmethod
    def convert_param_list_to_player_list(param_array : List[List[Param]]) -> PlayerList:
        player_list = PlayerList([])
        for element_array in param_array:
            # get element value with name "name"
            name = [param.value for param in element_array if param.name == "playerName"][0]

            player = Player(name)
            player_list.append(player)

        return player_list

    @staticmethod
    def convert_param_list_to_game_data(param_array : List[List[Param]]) -> GameData:
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

game_list_info = MessageParamListInfo(["gameName", "maxPlayers", "connectedPlayers"], Convertor.convert_param_list_to_game_list)
player_list_info = MessageParamListInfo(["playerName"], Convertor.convert_param_list_to_player_list)
game_data_info = MessageParamListInfo(["playerName", "isConnected", "score", "isTurn"], Convertor.convert_param_list_to_game_data)

class CCommandTypeEnum(Enum):
    # CLIENT->SERVER
    ClientLogin: Command = Command(1, GAME_STATE_MACHINE.ClientLogin, [], None)
    ClientCreateGame: Command = Command(2, GAME_STATE_MACHINE.ClientCreateGame, ["gameName", "maxPlayers"], None)
    ClientJoinGame: Command = Command(3, GAME_STATE_MACHINE.ClientJoinGame, ["gameName"], None)
    ClientStartGame: Command = Command(4, GAME_STATE_MACHINE.ClientStartGame, [], None)
    ClientRollDice: Command = Command(5, GAME_STATE_MACHINE.ClientRollDice, [], None)
    ClientLogout: Command = Command(7, GAME_STATE_MACHINE.ClientLogout, [], None)
    # ClientReconnect: Command = Command(8, G_game_state_machine.ClientReconnect, [], None)
    ClientNextDice: Command = Command(61, GAME_STATE_MACHINE.ClientNextDice, [], None)
    ClientEndTurn: Command = Command(62, GAME_STATE_MACHINE.ClientEndTurn, [], None)

    # RESPONSES SERVER->CLIENT
    ResponseServerSuccess: Command = Command(30, None, [], None)
    ResponseServerError: Command = Command(32, None, ["message"], None)

    ResponseServerGameList: Command = Command(33, None, ["gameList"], game_list_info)

    ResponseServerDiceNext: Command = Command(34, GAME_STATE_MACHINE.ResponseServerDiceNext, [], None)
    ResponseServerDiceEndTurn: Command = Command(35, GAME_STATE_MACHINE.ResponseServerDiceEndTurn, [], None)

    ResponseServerNextDiceEndScore: Command = Command(36, GAME_STATE_MACHINE.ResponseServerNextDiceEndScore, [], None)
    ResponseServerNextDiceSuccess: Command = Command(37, GAME_STATE_MACHINE.ResponseServerNextDiceSuccess, [], None)

    # SERVER->CLIENT
    ## SERVER -> ALL CLIENTS
    ServerUpdateStartGame: Command = Command(41, GAME_STATE_MACHINE.ServerUpdateStartGame, [], None)
    ServerUpdateEndScore: Command = Command(42, GAME_STATE_MACHINE.ServerUpdateEndScore, [], None)
    ServerUpdateGameData: Command = Command(43, GAME_STATE_MACHINE.ServerUpdateGameData, ["gameData"], game_data_info)
    ServerUpdateGameList: Command = Command(44, GAME_STATE_MACHINE.ServerUpdateGameList, ["gameList"], game_list_info)
    ServerUpdatePlayerList: Command = Command(45, GAME_STATE_MACHINE.ServerUpdatePlayerList, ["playerList"], player_list_info)

    ## SERVER -> SINGLE CLIENT
    ServerReconnectGameList: Command = Command(46, GAME_STATE_MACHINE.ServerReconnectGameList, ["gameList"], game_list_info)
    ServerReconnectGameData: Command = Command(47, GAME_STATE_MACHINE.ServerReconnectGameData, ["gameData"], game_data_info)
    ServerReconnectPlayerList: Command = Command(48, GAME_STATE_MACHINE.ServerReconnectPlayerList, ["playerList"], player_list_info)

    ServerStartTurn: Command = Command(49, GAME_STATE_MACHINE.ServerStartTurn, [], None)
    ServerPingPlayer: Command = Command(50, GAME_STATE_MACHINE.ServerPingPlayer, [], None)

    # RESPONSES CLIENT->SERVER
    ResponseClientSuccess: Command = Command(60, None, [], None)

    ErrorPlayerUnreachable: Command = Command(70, GAME_STATE_MACHINE.ErrorPlayerUnreachable, [], None)

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
    def is_valid_name(player_nickname: str) -> bool:
        return CMessageConfig.NAME_MIN_CHARS <= len(player_nickname) <= CMessageConfig.NAME_MAX_CHARS

@dataclass(frozen=True)
class CNetworkConfig:
    RECEIVE_TIMEOUT = 10
    BUFFER_SIZE: Final = 1024
    MAX_MESSAGE_SIZE: Final = 1024

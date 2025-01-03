from dataclasses import dataclass
from enum import Enum
from typing import Final

from backend.parser import Convertor
from backend.state_manager import GameStateMachine
from shared.data_structures import ScoreCube, Command, Brackets, MessageParamListInfo

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

game_list_info = MessageParamListInfo(["gameName", "maxPlayers", "connectedPlayers"], Convertor.convert_param_list_to_game_list)
player_list_info = MessageParamListInfo(["playerName"], Convertor.convert_param_list_to_player_list)
game_data_info = MessageParamListInfo(["playerName", "isConnected", "score", "isTurn"], Convertor.convert_param_list_to_game_data)

class CCommandTypeEnum(Enum):
    # CLIENT->SERVER
    ClientLogin: Command = Command(1, GAME_STATE_MACHINE.ClientLogin, [""], None)
    ClientCreateGame: Command = Command(2, GAME_STATE_MACHINE.ClientCreateGame, ["gameName, maxPlayers"], None)
    ClientJoinGame: Command = Command(3, GAME_STATE_MACHINE.ClientJoinGame, ["gameID"], None)
    ClientStartGame: Command = Command(4, GAME_STATE_MACHINE.ClientStartGame, [""], None)
    ClientRollDice: Command = Command(5, GAME_STATE_MACHINE.ClientRollDice, [""], None)
    ClientLogout: Command = Command(7, GAME_STATE_MACHINE.ClientLogout, [""], None)
    # ClientReconnect: Command = Command(8, G_game_state_machine.ClientReconnect, [""], None)
    ClientNextDice: Command = Command(61, GAME_STATE_MACHINE.ClientNextDice, [""], None)
    ClientEndTurn: Command = Command(62, GAME_STATE_MACHINE.ClientEndTurn, [""], None)

    # RESPONSES SERVER->CLIENT
    ResponseServerSuccess: Command = Command(30, None, [""], None)
    ResponseServerError: Command = Command(32, None, ["message"], None)
    ResponseServerGameList: Command = Command(33, None, ["gameList"], game_list_info)

    ResponseServerDiceNext: Command = Command(34, GAME_STATE_MACHINE.ResponseServerDiceNext, [""], None)
    ResponseServerDiceEndTurn: Command = Command(35, GAME_STATE_MACHINE.ResponseServerDiceEndTurn, [""], None)

    ResponseServerNextDiceEndScore: Command = Command(36, GAME_STATE_MACHINE.ResponseServerNextDiceEndScore, [""], None)
    ResponseServerNextDiceSuccess: Command = Command(37, GAME_STATE_MACHINE.ResponseServerNextDiceSuccess, [""], None)

    # SERVER->CLIENT
    ## SERVER -> ALL CLIENTS
    ServerUpdateStartGame: Command = Command(41, GAME_STATE_MACHINE.ServerUpdateStartGame, [""], None)
    ServerUpdateEndScore: Command = Command(42, GAME_STATE_MACHINE.ServerUpdateEndScore, [""], None)
    ServerUpdateGameData: Command = Command(43, GAME_STATE_MACHINE.ServerUpdateGameData, ["gameData"], game_data_info)
    ServerUpdateGameList: Command = Command(44, GAME_STATE_MACHINE.ServerUpdateGameList, ["gameList"], game_list_info)
    ServerUpdatePlayerList: Command = Command(45, GAME_STATE_MACHINE.ServerUpdatePlayerList, ["playerList"], player_list_info)

    ## SERVER -> SINGLE CLIENT
    ServerReconnectGameList: Command = Command(46, GAME_STATE_MACHINE.ServerReconnectGameList, ["gameList"], game_list_info)
    ServerReconnectGameData: Command = Command(47, GAME_STATE_MACHINE.ServerReconnectGameData, ["gameData"], game_data_info)
    ServerReconnectPlayerList: Command = Command(48, GAME_STATE_MACHINE.ServerReconnectPlayerList, ["playerList"], player_list_info)

    ServerStartTurn: Command = Command(49, GAME_STATE_MACHINE.ServerStartTurn, [""], None)
    ServerPingPlayer: Command = Command(50, GAME_STATE_MACHINE.ServerPingPlayer, [""], None)

    # RESPONSES CLIENT->SERVER
    # ResponseClientSuccess: Command = Command(60, G_game_state_machine.ResponseClientSuccess, [""], None)

    ErrorPlayerUnreachable: Command = Command(70, GAME_STATE_MACHINE.ErrorPlayerUnreachable, [""], None)

    @staticmethod
    def is_command_id_in_enum(command_id: int) -> bool:
        for command in CCommandTypeEnum:
            if command.value.id == command_id:
                return True
        return False

    @classmethod
    def get_command_by_id(cls, _command_id):
        for command in cls:
            if command.value.id == _command_id:
                return command
        return None


@dataclass(frozen=True)
class CMessagePartsSizes:
    TIMESTAMP_SIZE: Final = 32
    COMMANDID_SIZE: Final = 2
    SIGNATURE_SIZE: Final = 6


@dataclass(frozen=True)
class CMessageConfig:
    PARAMS_BRACKETS: Brackets = Brackets("{", "}")
    ARRAY_BRACKETS: Brackets = Brackets("[", "]")
    PARAMS_DELIMITER: str = ","
    PARAMS_KEY_VALUE_DELIMITER: str = ":"
    PARAMS_WRAPPER: str = "\""
    TIMESTAMP_FORMAT: str = '%Y-%m-%d %H:%M:%S'
    SIGNATURE: str = "KIVUPS"
    END_OF_MESSAGE: str = "\n"
    NAME_MIN_CHARS : Final = 3
    NAME_MAX_CHARS : Final = 20

    @staticmethod
    def is_valid_name(player_nickname: str) -> bool:
        return CMessageConfig.NAME_MIN_CHARS <= len(player_nickname) <= CMessageConfig.NAME_MAX_CHARS

@dataclass(frozen=True)
class CNetworkConfig:
    RECEIVE_TIMEOUT = 1
    BUFFER_SIZE: Final = 1024
    MAX_MESSAGE_SIZE: Final = 1024




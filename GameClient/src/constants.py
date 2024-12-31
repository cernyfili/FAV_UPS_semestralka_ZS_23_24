from enum import Enum
from state_machine import G_game_state_machine


class ScoreCube:
    def __init__(self, value, score_value):
        self.Value = value
        self.ScoreValue = score_value


CGMessageSignature = "KIVUPS"

CGNetworkEmptyParams = []

CGScoreCubeValues = [
    ScoreCube(1, 100),
    ScoreCube(5, 50),
]


class Command:
    def __init__(self, commandID, trigger, parameters):
        self.id = commandID
        self.trigger = trigger
        self.parameters = parameters


class CommandTypeEnum(Enum):
    # CLIENT->SERVER
    ClientLogin = Command(1, G_game_state_machine.ClientLogin, [""]),
    ClientCreateGame = Command(2, G_game_state_machine.ClientCreateGame, ["gameName, maxPlayers"]),
    ClientJoinGame = Command(3, G_game_state_machine.ClientJoinGame, ["gameID"]),
    ClientStartGame = Command(4, G_game_state_machine.ClientStartGame, [""]),
    ClientRollDice = Command(5, G_game_state_machine.ClientRollDice, [""]),
    ClientLogout = Command(7, G_game_state_machine.ClientLogout, [""]),
    ClientReconnect = Command(8, G_game_state_machine.ClientReconnect, [""]),
    ClientNextDice = Command(61, G_game_state_machine.ClientNextDice, [""]),
    ClientEndTurn = Command(62, G_game_state_machine.ClientEndTurn, [""]),

    # RESPONSES SERVER->CLIENT
    ResponseServerSuccess = Command(30, None, [""]),
    ResponseServerErrDuplicitNickname = Command(31, None, [""]),
    ResponseServerError = Command(32, None, ["message"]),
    ResponseServerGameList = Command(33, None, ["gameList"]),

    ResponseServerDiceNext = Command(34, G_game_state_machine.ResponseServerDiceNext, [""]),
    ResponseServerDiceEndTurn = Command(35, G_game_state_machine.ResponseServerDiceEndTurn, [""]),

    ResponseServerNextDiceEndScore = Command(36, G_game_state_machine.ResponseServerNextDiceEndScore, [""]),
    ResponseServerNextDiceSuccess = Command(37, G_game_state_machine.ResponseServerNextDiceSuccess, [""]),

    # SERVER->CLIENT
    ServerUpdateStartGame = Command(41, G_game_state_machine.ServerUpdateStartGame, [""]),
    ServerUpdateEndScore = Command(42, G_game_state_machine.ServerUpdateEndScore, [""]),
    ServerUpdateGameData = Command(43, G_game_state_machine.ServerUpdateGameData, [""]),
    ServerUpdateGameList = Command(44, G_game_state_machine.ServerUpdateGameList, ["gameList"]),
    ServerUpdatePlayerList = Command(45, G_game_state_machine.ServerUpdatePlayerList, [""]),

    ServerReconnectGameList = Command(46, G_game_state_machine.ServerReconnectGameList, [""]),
    ServerReconnectGameData = Command(47, G_game_state_machine.ServerReconnectGameData, [""]),
    ServerReconnectPlayerList = Command(48, G_game_state_machine.ServerReconnectPlayerList, [""]),

    ServerStartTurn = Command(49, G_game_state_machine.ServerStartTurn, [""]),
    ServerPingPlayer = Command(50, G_game_state_machine.ServerPingPlayer, [""]),

    # RESPONSES CLIENT->SERVER
    ResponseClientSuccess = Command(60, G_game_state_machine.ResponseClientSuccess, [""]),

    ErrorPlayerUnreachable = Command(70, G_game_state_machine.ErrorPlayerUnreachable, [""]),

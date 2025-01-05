import logging
import socket
import threading
from datetime import datetime
from typing import List

from backend.parser import convert_message_to_network_string, parse_message
from shared.constants import CMessageConfig, CCommandTypeEnum, \
    CNetworkConfig, GAME_STATE_MACHINE, Command, Param, NetworkMessage, GameList, reset_game_state_machine, PlayerList


#
# def send_and_receive(server_ip, server_port, message):
#     # Create a socket object
#     s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#
#     # Connect to the server
#     s.connect((server_ip, server_port))
#
#     # Send data to the server
#     s.sendall(message.encode())
#
#     # Receive data from the server
#     data = s.recv(1024)
#
#     # Close the connection
#     s.close()
#
#     return data.decode()


class ServerCommunication:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(ServerCommunication, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        #self._can_receive = False
        if self._initialized:
            return
        self._ip = None
        self._port = None
        self._s = None
        self._nickname = None
        self._initialized = True
        self._lock = threading.Lock()  # Initialize the lock

        self._send_messages_list : List[NetworkMessage] = []
        self._receive_messages_list : List[NetworkMessage] = []

    # def _send_and_receive(self, command_id : int, nickname : str, param : List[Param]) -> str:
    #     if self._s is None:
    #         raise ConnectionError("Server is not connected.")
    #     message = self.create_message(command_id, param)
    #     self._s.sendall(message.encode())
    #     data = self._s.recv(1024)
    #     return data.decode()

    def _close_connection(self):
        if self._s is not None:
            self._s.close()
        self._s = None

    def _connect_to_server(self, ip : str, port : int):
        # Create a socket object
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

        # Connect to the server
        s.connect((ip, port))

        self._ip = ip
        self._port = port
        self._s = s

    def _send_message(self, command : Command, param=None) -> None:
        def create_message(command : Command, param : List[Param]) -> NetworkMessage:
            command_id = command.id
            nickname = self._nickname
            timestamp = datetime.now().strftime(CMessageConfig.TIMESTAMP_FORMAT)

            message = NetworkMessage(CMessageConfig.SIGNATURE, command_id, timestamp, nickname, param)
            return message

        if not isinstance(param, List) and param is not None:
            raise ValueError("Invalid parameter type.")

        if param is None:
            param = []

        message = create_message(command, param)

        message_str = convert_message_to_network_string(message)
        # check if right param

        self._s.sendall(message_str.encode())
        self._send_messages_list.append(message)
        logging.debug(f"Sent message: {message}")

    def is_header_valid(self, message : NetworkMessage) -> bool:
        if message.signature != CMessageConfig.SIGNATURE:
            return False
        if message.command_id is None or not CCommandTypeEnum.is_command_id_in_enum(message.command_id):
            return False
        if message.timestamp is None:
            return False
        if message.player_nickname is None or message.player_nickname != self._nickname:
            return False
        return True

    # region receive message
    def _receive_message(self) -> NetworkMessage:
        with self._lock:
            self._s.settimeout(CNetworkConfig.RECEIVE_TIMEOUT)
            data = b""
            try:
                data = self.__receive_loop(data)
            except socket.timeout:
                raise ConnectionError("Server did not respond in time.")
            finally:
                self._s.settimeout(None)  # Reset timeout to default

            parsed_message = self.__process_received_data(data)
            return parsed_message

    def _receive_message_without_timeout(self):
        data = b""


        part = self._s.recv(CNetworkConfig.BUFFER_SIZE)
        logging.debug(f"Received part: {part}")
        data += part
        if CMessageConfig.END_OF_MESSAGE.encode() in part or len(data) >= CNetworkConfig.MAX_MESSAGE_SIZE:
            parsed_message = self.__process_received_data(data)
            return parsed_message
        return None



    def __process_received_data(self, data):
        parsed_message = parse_message(data.decode())
        if not self.is_header_valid(parsed_message):
            raise ValueError("Invalid header in the received message.")
        self._receive_messages_list.append(parsed_message)
        logging.debug(f"Received message: {parsed_message}")
        return parsed_message

    def __receive_loop(self, data):
        while True:
            part = self._s.recv(CNetworkConfig.BUFFER_SIZE)
            logging.debug(f"Received part: {part}")
            data += part
            if CMessageConfig.END_OF_MESSAGE.encode() in part or len(data) >= CNetworkConfig.MAX_MESSAGE_SIZE:
                break
        return data
    # endregion

    @staticmethod
    def convert_params_error_message(parameters: List[Param]) -> str:
        # Convert the parameters to a string
        if len(parameters) != 1:
            error_message = ""
        error_message = parameters[0].value

        return error_message

    def _process_wrong_message(self):
        self._close_connection()
        reset_game_state_machine()

    # region COMMUNICATION FUNCTIONS

    # region SEND

    def _send_standard_command(self, command : Command, allowed_response_commands_id, param_list) -> bool:
        #self._can_receive = False
        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC

        # region SEND
        self._send_message(command, param_list)
        # endregion

        # region RESPONSE
        #self._can_receive = True
        received_message = self._receive_message()
        # Response other
        if received_message.command_id not in allowed_response_commands_id:
            self._process_wrong_message()
            return False

        # Response Error
        if received_message.command_id == CCommandTypeEnum.ResponseServerError.value.id:
            error_message = self.convert_params_error_message(received_message.parameters)
            raise ConnectionError(f"Error: {error_message}")

        # Response Success
        if received_message.command_id != CCommandTypeEnum.ResponseServerSuccess.value.id:
            raise ValueError("Invalid command ID for game list update.")

        # endregion
        GAME_STATE_MACHINE.send_trigger(command.trigger)
        return True

    def send_client_login(self, ip, port, nickname) -> tuple[bool, GameList | None]:
        command = CCommandTypeEnum.ClientLogin.value

        allowed_response_commands_id = [CCommandTypeEnum.ResponseServerGameList.value.id, CCommandTypeEnum.ResponseServerError.value.id]

        #self._can_receive = False
        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC
        self._nickname = nickname
        # region SEND
        # Connect to the server
        self._connect_to_server(ip, port)
        self._send_message(command)
        # endregion

        # region RESPONSE
        #self._can_receive = True
        try:
            received_message = self._receive_message()
        except Exception as e:
        # NOT RECEIVED MESSAGE IN TIMEOUT -> DISCONNECT
            self._close_connection()
            raise e

        # region RECEIVED MESSAGE
        # received_message other
        if received_message.command_id not in allowed_response_commands_id:
            self._process_wrong_message()
            return False, None

        # received_message = Error
        if received_message.command_id == CCommandTypeEnum.ResponseServerError.value.id:
            error_message = self.convert_params_error_message(received_message.parameters)
            self._close_connection()
            raise ConnectionError(f"Error: {error_message}")

        # received_message = Success (GameList)
        if received_message.command_id != CCommandTypeEnum.ResponseServerGameList.value.id:
            raise ValueError("Invalid command ID for game list update.")

        return_value = None if not received_message.parameters else received_message.parameters[0].value
        # endregion

        # endregion

        # endregion

        GAME_STATE_MACHINE.send_trigger(command.trigger)
        return True, return_value


    def send_client_create_game(self, game_name, max_players_count) -> bool:
        command = CCommandTypeEnum.ClientCreateGame.value
        allowed_response_commands = [CCommandTypeEnum.ResponseServerSuccess.value.id, CCommandTypeEnum.ResponseServerError.value.id]
        param_list = [Param("gameName", game_name), Param("maxPlayers", max_players_count)]

        return self._send_standard_command(command, allowed_response_commands, param_list)

    def send_client_join_game(self, game_name) -> bool:
        command = CCommandTypeEnum.ClientJoinGame.value
        allowed_response_commands = [CCommandTypeEnum.ResponseServerSuccess.value.id, CCommandTypeEnum.ResponseServerError.value.id]
        param_list = [Param("gameName", game_name)]

        return self._send_standard_command(command, allowed_response_commands, param_list)

    def send_start_game(self) -> bool:
        command = CCommandTypeEnum.ClientStartGame.value
        allowed_response_commands = [CCommandTypeEnum.ResponseServerSuccess.value.id, CCommandTypeEnum.ResponseServerError.value.id]

        return self._send_standard_command(command, allowed_response_commands, [])
    # endregion

    # region RECEIVE
    def _receive_standard_command(self, command : Command):
        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC
        try:
            received_message = self._receive_message()
        except Exception as e:
            return None

        if received_message is None:
            return None
        if received_message.command_id != command.id:
            return None
            #raise ValueError("Invalid command ID for game list update.")

        return_value = received_message.get_array_param()

        # region SEND RESPONSE

        self._respond_client_success()

        # endregion

        # endregion

        GAME_STATE_MACHINE.send_trigger(command.trigger)
        return return_value

    def receive_game_list_update(self) -> GameList | None:
        command = CCommandTypeEnum.ServerUpdateGameList.value
        return self._receive_standard_command(command)

    def receive_player_list_update(self) -> PlayerList | None:
        command = CCommandTypeEnum.ServerUpdatePlayerList.value
        return self._receive_standard_command(command)

    # endregion

    # region RESPONSE

    def _respond_client_success(self):
        command = CCommandTypeEnum.ResponseClientSuccess.value
        self._send_message(command)

    # endregion

    # get nickname
    @property
    def nickname(self):
        return self._nickname






import socket
from datetime import datetime
from typing import List

from backend.parser import convert_message_to_network_string, parse_message
from shared.constants import MESSAGE_SIGNATURE, CMessageConfig, CCommandTypeEnum, \
    CNetworkConfig, GAME_STATE_MACHINE
from shared.data_structures import Param, Command, NetworkMessage


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
        if self._initialized:
            return
        self._ip = None
        self._port = None
        self._s = None
        self._nickname = None
        self._initialized = True

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

            # check if right param
            if not command.is_valid_param_names(param):
                raise ValueError("Invalid parameter names.")

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

    def _receive_message(self) -> NetworkMessage:
        self._s.settimeout(CNetworkConfig.RECEIVE_TIMEOUT)
        data = b""
        try:
            while True:
                part = self._s.recv(CNetworkConfig.BUFFER_SIZE)
                data += part
                if CMessageConfig.END_OF_MESSAGE.encode() in part or len(data) >= CNetworkConfig.MAX_MESSAGE_SIZE:
                    break
        except socket.timeout:
            raise ConnectionError("Server did not respond in time.")
        finally:
            self._s.settimeout(None)  # Reset timeout to default

        parsed_message = parse_message(data.decode())
        self._receive_messages_list.append(parsed_message)
        return parsed_message

    @staticmethod
    def convert_params_error_message(parameters: List[Param]) -> str:
        # Convert the parameters to a string
        if len(parameters) != 1:
            error_message = ""
        error_message = parameters[0].value

        return error_message

    def send_client_login(self, ip, port, nickname):
        command = CCommandTypeEnum.ClientLogin.value

        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can(command.trigger):
            raise ValueError("Invalid state machine transition.")

        self._nickname = nickname
        # Connect to the server
        try:
            self._connect_to_server(ip, port)
            self._send_message(command)
        except all:
            # Show an error message if the connection fails
            raise ConnectionError("Unable to connect to the server!")

        # region RESPONSE
        try:
            received_message = self._receive_message()
        except Exception as e:
        # NOT RECEIVED MESSAGE IN TIMEOUT -> DISCONNECT
            self._close_connection()
            raise e

        # region RECEIVED MESSAGE
        # received_message = Error
        if received_message.command_id == CCommandTypeEnum.ResponseServerError.value.id:
            error_message = self.convert_params_error_message(received_message.parameters)
            self._close_connection()
            raise ConnectionError(f"Error: {error_message}")

        # received_message = Success (GameList)
        if received_message.command_id == CCommandTypeEnum.ResponseServerGameList.value.id:
            game_list = received_message.parameters[0].value
            GAME_STATE_MACHINE.send(command.trigger)
            return game_list

        # received_message other
        self._close_connection()
        raise ValueError("Invalid response from the server.")
        # endregion

        # endregion



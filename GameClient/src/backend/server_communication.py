import logging
import socket
import sys
import threading
import time
from datetime import datetime
from typing import List

from src.backend.parser import convert_message_to_network_string, parse_message, \
    convert_list_cube_values_to_network_string
from src.shared.constants import CMessageConfig, CCommandTypeEnum, \
    CNetworkConfig, GAME_STATE_MACHINE, Command, Param, NetworkMessage, GameList, reset_game_state_machine, PlayerList, \
    CubeValuesList, GameData


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
        # self._can_receive = False
        if self._initialized:
            return
        self._ip = None
        self._port = None
        self._s = None
        self._nickname = None
        self._initialized = True
        self._was_connected = False

        self._lock = threading.Lock()  # Initialize the lock

        self._send_messages_list: List[NetworkMessage] = []
        self._receive_messages_list: List[NetworkMessage] = []

        self._received_messages_to_process_list: List[NetworkMessage] = []

    # region PRIVATE FUNCTIONS


    def _connect_to_server(self, ip: str, port: int):
        # Create a socket object
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

        # Connect to the server
        s.connect((ip, port))

        self._ip = ip
        self._port = port
        self._s = s

        self._was_connected = True

    def _send_message(self, command: Command, param=None) -> bool:
        """

        :param command:
        :param param:
        :return: bool - isConnected
        """

        def create_message(command: Command, param: List[Param]) -> NetworkMessage:
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

        try:
            self._s.sendall(message_str.encode())
            self._send_messages_list.append(message)
            logging.info(f"MESSAGE: Sent message: {message}")
            logging.debug(f"Sent message string: {message_str}")
            return True
        except Exception as e:
            logging.error(f"Failed to send message: {e}")
            self._close_connection_processes()
            return False

    def _is_header_valid(self, message: NetworkMessage) -> bool:
        if message.signature != CMessageConfig.SIGNATURE:
            return False
        if message.command_id is None or not CCommandTypeEnum.is_command_id_in_enum(message.command_id):
            return False
        if message.timestamp is None:
            return False
        if message.player_nickname is None or message.player_nickname != self._nickname:
            return False
        return True

    def process_reconnect(self) -> bool:
        """
        :param self:
        :return:
        is_connected: bool
        """
        def __reconnect(self):
            logging.info("Attempting to reconnect...")
            self._connect_to_server(self._ip, self._port)

        max_retries = CNetworkConfig.RECONNECT_ATTEMPTS
        wait_time = CNetworkConfig.RECONNECT_TIMEOUT_SEC

        for attempt in range(max_retries):
            time.sleep(wait_time)
            try:
                __reconnect(self)
                return True
            except Exception as e:
                continue

        # reconnect failed
        self._close_connection_processes()
        return False

    # region receive message
    def _receive_message(self) -> tuple[bool, NetworkMessage | None]:
        """
            :return:
                bool - is connected
                NetworkMessage - received message
        """


        def __receive_one_messages(self) -> tuple[bool, NetworkMessage | None]:
            """
            :param self:
            :return:
             bool - is connected
            """

            # if is in process list something return this
            if self._received_messages_to_process_list:
                received_message = self._received_messages_to_process_list.pop(0)
                return True, received_message

            data = b""

            try:
                data = self._receive_loop(data)
                if not data:
                    self._close_connection_processes()
                    return False, None
            except Exception as e:
                # exception -> closed connection from server - trying to recconect
                # exception -> timeout
                logging.error(f"Server did not respond in time. {e}")
                self._close_connection_processes()
                return False, None

            is_connected, parsed_message_list = self._process_received_data(data)
            if len(parsed_message_list) != 1:
                self._received_messages_to_process_list.extend(parsed_message_list[1:])

            received_message = parsed_message_list[0]
            return is_connected, received_message

        with self._lock:
            while True:
                is_connected, received_message = __receive_one_messages(self)
                if not is_connected:
                    return False, None

                # process ServerPingPlayer
                received_command = CCommandTypeEnum.get_command_by_id(received_message.command_id)
                if received_command.id == CCommandTypeEnum.ServerPingPlayer.value.id:
                    is_connected = self._process_respond_successs(received_command)
                    if not is_connected:
                        return False, None
                    continue

                break

            return is_connected, received_message
    # def receive_messages(self):
    #     self._s.settimeout(CNetworkConfig.RECEIVE_TIMEOUT)
    #     data = b""
    #     max_retries = CNetworkConfig.RECONNECT_ATTEMPTS
    #     for attempt in range(max_retries):
    #         try:
    #             data = self._receive_loop(data)
    #             if not data:
    #                 # no received data -> closed connection from server - trying to recconect
    #                 is_connected = self._process_reconnect(attempt, max_retries)
    #                 if not is_connected:
    #                     return False, None
    #                 continue
    #             break
    #         except Exception as e:
    #             # exception -> closed connection from server - trying to recconect
    #             logging.error("Server did not respond in time.")
    #             is_connected = self._process_reconnect(attempt, max_retries)
    #             if not is_connected:
    #                 return False, None
    #             continue
    #         finally:
    #             self._s.settimeout(None)  # Reset timeout to default
    #     is_connected, parsed_message_list = self._process_received_data(data)
    #     if len(parsed_message_list) != 1:
    #         raise ValueError("Invalid number of messages received.")
    #     received_message = parsed_message_list[0]
    #     return is_connected, parsed_message_list, received_message
    def _receive_message_list(self) -> tuple[bool, list[NetworkMessage] | None]:
        """
        :return:
            bool - is connected
            bool - is timeout
            NetworkMessage - received message
        """
        logging.debug("Receiving message without reconnect...")
        with self._lock:
            data = b""
            max_retries = CNetworkConfig.RECONNECT_ATTEMPTS
            for attempt in range(max_retries):
                try:
                    data = self._receive_loop(data)
                    if not data:
                        self._close_connection_processes()
                        return False, None
                    break
                # except socket.timeout:
                #     is_connected = True
                #     is_timeout = True
                #     return is_connected, is_timeout, None
                except Exception as e:
                    logging.error("Server did not respond in time.")
                    self._close_connection_processes()
                    return False, None
                    continue

            is_connected, parsed_message_list = self._process_received_data(data)
            return is_connected, parsed_message_list

    def _process_received_data(self, received_data_list) -> tuple[bool, list[NetworkMessage] | None]:
        received_data = received_data_list.decode()
        if len(received_data) == 0:
            raise ConnectionError("Received empty message.")
        # split to array by end of message
        received_data_list = received_data.split(CMessageConfig.END_OF_MESSAGE)

        parsed_messages_list = []
        for data in received_data_list:
            if data == "":
                continue
            try:
                parsed_message = parse_message(data)
                is_valid_format = self._is_header_valid(parsed_message)
                if not is_valid_format:
                    # is not valid format
                    self._close_connection_processes()
                    is_connected = False
                    return is_connected, None
                parsed_messages_list.append(parsed_message)
                self._receive_messages_list.append(parsed_message)
                logging.info(f"MESSAGE: Received message: {parsed_message}")
            except Exception as e:
                # is not valid format
                self._close_connection_processes()
                is_connected = False
                return is_connected, None

        is_connected = True
        return is_connected, parsed_messages_list

    def _receive_loop(self, data):
        self._s.settimeout(CNetworkConfig.RECEIVE_TIMEOUT)

        logging.debug(f"Receiving data: {data}")
        while True:
            try:
                part = self._s.recv(CNetworkConfig.BUFFER_SIZE)
            except Exception as e:
                raise e
            finally:
                self._s.settimeout(None)

            if part != b"":
                logging.debug(f"Received part: {part}")
            if not part:
                break
            # logging.debug(f"Received part: {part}")

            data += part
            decoded_data = data.decode()
            if CMessageConfig.END_OF_MESSAGE in decoded_data or len(data) >= CNetworkConfig.MAX_MESSAGE_SIZE:
                logging.debug(f"----Received data: {data}")
                break

        self._s.settimeout(None)
        return data

    # endregion

    @staticmethod
    def _convert_params_error_message(parameters: List[Param]) -> str:
        # Convert the parameters to a string
        if len(parameters) != 1:
            error_message = ""
        error_message = parameters[0].value

        return error_message

    def _close_connection_processes(self):
        def _close_connection(self):
            logging.info("Closing connection...")
            if self._s is not None:
                try:
                    self._s.shutdown(socket.SHUT_RDWR)
                    self._s.close()
                    logging.info("Connection closed successfully.")
                except Exception as e:
                    logging.error(f"Failed to close connection: {e}")
                finally:
                    self._s = None

        reset_game_state_machine()
        _close_connection(self)

    # endregion

    # region COMMUNICATION FUNCTIONS

    # region SEND

    # region PRIVATE FUNCTIONS
    @staticmethod
    def _state_machine_send_triggers(send_command, received_command):
        GAME_STATE_MACHINE.send_trigger(send_command.trigger)
        GAME_STATE_MACHINE.send_trigger(received_command.trigger)

    def _send_standard_command(self, command: Command, param_list) -> bool:
        """

        :param command:
        :param param_list:
        :return: bool - isConnected
        """
        def __disconnect(self):
            self._close_connection_processes()
            return False

        allowed_response_commands_id = [CCommandTypeEnum.ResponseServerError.value.id,
                                        CCommandTypeEnum.ResponseServerSuccess.value.id]

        # self._can_receive = False
        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC

        # region SEND
        is_connected = self._send_message(command, param_list)
        if not is_connected:
            return False
        # endregion

        # region RESPONSE
        # self._can_receive = True
        is_connected, received_message = self._receive_message()
        if not is_connected:
            return False

        # Response other
        if received_message.command_id not in allowed_response_commands_id:
            logging.error(f"Invalid command ID received: {received_message.command_id} but allowed: {allowed_response_commands_id}.")
            return __disconnect(self)

        # Response Error
        self._process_error_message(command, received_message)

        # Response Success
        if received_message.command_id != CCommandTypeEnum.ResponseServerSuccess.value.id:
            raise ValueError("Invalid command ID for game list update.")

        # endregion
        GAME_STATE_MACHINE.send_trigger(command.trigger)
        return True

    def _process_error_message(self, received_command, received_message):
        if received_command.id == CCommandTypeEnum.ResponseServerError.value.id:
            error_message = self._convert_params_error_message(received_message.parameters)
            raise ConnectionError(f"Error: {error_message}")

    # endregion

    def _receive_connect_message(self, command : Command, allowed_response_commands_id, process_response_commands_id_list):
        def __disconnect(self):
            self._close_connection_processes()
            return False, None, None

        # region RESPONSE
        # self._can_receive = True
        try:
            is_connected, received_message = self._receive_message()
            if not is_connected:
                return False, None, None
        except Exception as e:
            # NOT RECEIVED MESSAGE IN TIMEOUT -> DISCONNECT
            return __disconnect(self)

        # region RECEIVED MESSAGE

        # if not in allowed_response_commands_id
        if received_message.command_id not in allowed_response_commands_id:
            return __disconnect(self)

        # received_message = Error
        if received_message.command_id == CCommandTypeEnum.ResponseServerError.value.id:
            error_message = self._convert_params_error_message(received_message.parameters)
            self._close_connection_processes()
            raise ConnectionError(f"Error: {error_message}")

        # received_message = Success
        for process_response_command_id in process_response_commands_id_list:
            if received_message.command_id != process_response_command_id:
                continue

            try:
                GAME_STATE_MACHINE.send_trigger(command.trigger)
            except Exception as e:
                return __disconnect(self)

            message_list = received_message.get_array_param()
            return True, received_message, message_list
        # endregion

        # endregion

        # unreacheable
        logging.error("Invalid command ID for connect message.")
        sys.exit(1)

    def send_connect_message(self, ip, port, nickname) -> tuple[bool, NetworkMessage | None, GameList | None]:


        if self._was_connected:
            if self._nickname != nickname:
                self._was_connected = False

        # if self._was_connected then ClientReconnect
        if self._was_connected:
            command = CCommandTypeEnum.ClientReconnect.value
            process_response_commands_id_list = [CCommandTypeEnum.ResponseServerReconnectBeforeGame.value.id,
                                                 CCommandTypeEnum.ResponseServerReconnectRunningGame.value.id,
                                                 CCommandTypeEnum.ResponseServerGameList.value.id
                                                 ]
        else:
            command = CCommandTypeEnum.ClientLogin.value
            process_response_commands_id_list = [CCommandTypeEnum.ResponseServerGameList.value.id]

        allowed_response_commands_id = process_response_commands_id_list + [
            CCommandTypeEnum.ResponseServerError.value.id]

        # self._can_receive = False
        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC
        self._nickname = nickname
        # region SEND
        # Connect to the server
        self._connect_to_server(ip, port)
        is_connected = self._send_message(command)
        if not is_connected:
            return False, None, None
        # endregion

        # region RESPONSE
        is_connected, received_message, message_list = self._receive_connect_message(command, allowed_response_commands_id,

                                                                                        process_response_commands_id_list)
        return is_connected, received_message, message_list

    def communication_reconnect_message(self) -> tuple[bool, NetworkMessage | None, list | None]:
        """

        :return:
        is_connected: bool
        received_message: NetworkMessage
        message_data: list
        """

        if self.is_connected:
            assert False, "Trying to reconnect when not disconnected"

        command : Command = CCommandTypeEnum.ClientReconnect.value
        process_response_commands_id_list = [CCommandTypeEnum.ResponseServerReconnectBeforeGame.value.id,
                                             CCommandTypeEnum.ResponseServerReconnectRunningGame.value.id,
                                             CCommandTypeEnum.ResponseServerGameList.value.id
                                             ]
        allowed_response_commands_id = process_response_commands_id_list + [
            CCommandTypeEnum.ResponseServerError.value.id]


        is_connected = self.process_reconnect()
        if not is_connected:
            return False, None, None

        #todo
        # if not GAME_STATE_MACHINE.can_fire(command.trigger):
        #     assert False, f"Invalid state machine transition for command: {command.id} with trigger: {command.trigger}."

        is_connected = self._send_message(command)
        if not is_connected:
            return False, None, None

        # Response
        is_connected, received_message, message_list = self._receive_connect_message(command, allowed_response_commands_id,
                                                                                    process_response_commands_id_list)
        return is_connected, received_message, message_list




    def send_client_create_game(self, param_list: list[Param]) -> bool:
        def __is_param_list_valid(command: Command, param_list: list[Param]) -> bool:
            if not command.is_valid_param_names(param_list):
                return False
            if param_list[0].name != "gameName" and isinstance(param_list[0].value, str):
                return False
            if param_list[1].name != "maxPlayers" and isinstance(param_list[1].value, int):
                return False

            return True

        command = CCommandTypeEnum.ClientCreateGame.value
        if not __is_param_list_valid(command, param_list):
            raise ValueError("Invalid parameter list for create game.")

        return self._send_standard_command(command, param_list)

    def send_client_join_game(self, param_list) -> bool:
        def __is_param_list_valid(command: Command, param_list: list[Param]) -> bool:
            if not command.is_valid_param_names(param_list):
                return False
            if param_list[0].name != "gameName" and isinstance(param_list[0].value, str):
                return False

            return True

        command = CCommandTypeEnum.ClientJoinGame.value

        if not __is_param_list_valid(command, param_list):
            raise ValueError("Invalid parameter list for join game.")

        return self._send_standard_command(command, param_list)

    def send_client_start_game(self, param_list: list[Param]) -> bool:
        def __is_param_list_valid(command: Command, param_list: list[Param]) -> bool:
            if not command.is_valid_param_names(param_list):
                return False

            return True

        command = CCommandTypeEnum.ClientStartGame.value

        if not __is_param_list_valid(command, param_list):
            raise ValueError("Invalid parameter list for start game.")

        return self._send_standard_command(command, param_list)

    def send_client_logout(self) -> bool:
        def __disconnect(self):
            self._close_connection_processes()
            return False

        command = CCommandTypeEnum.ClientLogout.value
        param_list = []

        allowed_response_commands_id = [CCommandTypeEnum.ResponseServerError.value.id,
                                        CCommandTypeEnum.ResponseServerSuccess.value.id]

        # region LOGIC

        # region SEND
        is_connected = self._send_message(command, param_list)
        if not is_connected:
            return False
        # endregion

        # region RESPONSE
        # self._can_receive = True
        is_connected, received_message = self._receive_message()

        if not is_connected:
            return False
        # Response other
        if received_message.command_id not in allowed_response_commands_id:
            return __disconnect(self)

        # Response Error
        self._process_error_message(command, received_message)

        # Response Success
        if received_message.command_id == CCommandTypeEnum.ResponseServerSuccess.value.id:
            return True

        raise ValueError("Invalid command ID for game list update.")

        # endregion

    def send_client_roll_dice(self) -> tuple[bool, Command | None, CubeValuesList | None]:
        def __disconnect(self):
            self._close_connection_processes()
            return False, None, None

        command = CCommandTypeEnum.ClientRollDice.value

        allowed_response_commands_id = [CCommandTypeEnum.ResponseServerSelectCubes.value.id,
                                        CCommandTypeEnum.ResponseServerEndTurn.value.id,
                                        CCommandTypeEnum.ResponseServerError.value.id]

        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC
        # region SEND
        # Connect to the server
        is_connected = self._send_message(command)
        if not is_connected:
            return False, None, None
        # endregion

        GAME_STATE_MACHINE.send_trigger(command.trigger)

        # region RESPONSE

        is_connected, received_message = self._receive_message()
        if not is_connected:
            return False, None, None
        if received_message.command_id == CCommandTypeEnum.ServerPingPlayer.value.id:
            is_connected, received_message = self._receive_message()
            if not is_connected:
                return False, None, None

        received_command = CCommandTypeEnum.get_command_by_id(received_message.command_id)

        # Response other
        if received_command.id not in allowed_response_commands_id:
            logging.error("Invalid command")
            return __disconnect(self)

        # Response Error
        self._process_error_message(received_command, received_message)

        # Response SUCCESS

        # Response SelectCubes
        expected_command = CCommandTypeEnum.ResponseServerSelectCubes.value
        if received_command.id == expected_command.id:
            received_command = expected_command
            GAME_STATE_MACHINE.send_trigger(received_command.trigger)
            return_value = received_message.get_array_param()
            return True, received_command, return_value

        # Response EndTurn
        expected_command = CCommandTypeEnum.ResponseServerEndTurn.value
        if received_command.id == expected_command.id:
            received_command = expected_command
            GAME_STATE_MACHINE.send_trigger(received_command.trigger)
            return True, received_command, None

        # unreacheble
        raise ValueError("Invalid command ID for roll dice.")
        # endregion

        # endregion

    def send_client_select_cubes(self, selected_cubes: list[int]) -> tuple[bool, Command | None]:
        def __disconnect(self):
            self._close_connection_processes()
            return False, None

        command = CCommandTypeEnum.ClientSelectedCubes.value

        success_response_command = [CCommandTypeEnum.ResponseServerDiceSuccess.value.id,
                                    CCommandTypeEnum.ResponseServerEndScore.value.id]
        allowed_response_commands_id = success_response_command + [CCommandTypeEnum.ResponseServerError.value.id]

        # check if statemachine can fire
        if not GAME_STATE_MACHINE.can_fire(command.trigger.id):
            raise ValueError("Invalid state machine transition.")

        # region LOGIC
        # region SEND
        # Connect to the server
        value_cube_value = convert_list_cube_values_to_network_string(selected_cubes)

        param_list = [Param("cubeValues", value_cube_value)]
        is_connected = self._send_message(command, param_list)
        if not is_connected:
            return False, None
        # endregion

        # region RESPONSE
        is_connected, received_message = self._receive_message()
        if not is_connected:
            return False, None
        if received_message.command_id == CCommandTypeEnum.ServerPingPlayer.value.id:
            is_connected, received_message = self._receive_message()
            if not is_connected:
                return False, None

        received_command = CCommandTypeEnum.get_command_by_id(received_message.command_id)

        # Response other
        if received_command.id not in allowed_response_commands_id:
            return __disconnect(self)

        # Response Error
        self._process_error_message(received_command, received_message)

        # Response SUCCESS

        if not received_command.id in success_response_command:
            raise ValueError("Invalid command ID for select cubes.")

        # Response Dice Success or EndScore
        self._state_machine_send_triggers(command, received_command)
        return True, received_command

        # endregion

        # endregion

    # endregion

    # region RECEIVE
    # region PRIVATE FUNCTIONS
    # def _receive_standard_update_command(self, command : Command) -> tuple[bool, any]:
    #     """
    #
    #     :param command:
    #     :return: list of updated values
    #     """
    #     is_connected, is_timeout, recieved_message = self.__receive_update_command(command)
    #     if is_timeout:
    #         return is_connected, None
    #     return is_connected, recieved_message.get_array_param()

    # def __receive_update_command(self, command: Command) -> tuple[bool, bool, NetworkMessage | None]:
    #     """
    #     :param command: Command object to be processed
    #     :return: tuple containing:
    #         - bool: is_connected
    #         - bool: is_timeout
    #         - NetworkMessage | None: the received message or None if not received
    #     """
    #
    #     def __disconnect(self):
    #         self._close_connection_processes()
    #         return False, False, None
    #
    #     # region LOGIC
    #     is_connected, is_timeout, received_message_list = self._receive_message_with_timeout_return()
    #     if is_timeout:
    #         return is_connected, is_timeout, None
    #
    #     for received_message in received_message_list:
    #
    #         received_command_id = received_message.command_id
    #         received_command = CCommandTypeEnum.get_command_by_id(received_command_id)
    #
    #         # check if statemachine can fire
    #         if not GAME_STATE_MACHINE.can_fire(received_command.trigger.id):
    #             raise ValueError("Invalid state machine transition.")
    #
    #         # Recieved Other
    #         if received_command_id != command.id:
    #             logging.error("Invalid command")
    #             return __disconnect(self)
    #
    #         # region SEND RESPONSE
    #
    #         self._respond_client_success()
    #
    #         # endregion
    #
    #         # endregion
    #
    #         GAME_STATE_MACHINE.send_trigger(received_command.trigger)
    #
    #     #received_message is the last if received_message_list is not len 0
    #     return_received_message = received_message_list[-1] if received_message_list else None
    #
    #     return is_connected, is_timeout, return_received_message

    def _process_server_ping_player(self, received_command, received_message) -> tuple[Command | None, None]:
        is_connected = self._process_respond_successs(received_command)
        if not is_connected:
            return None, None

        return received_command, None

    def _receive_standard_state_messages(self, allowed_commands: dict[int, callable]) -> tuple[
        bool, list[tuple[Command | None, any]] | None]:
        """

        :param allowed_commands:
        :return:
        bool - is connected
        list of tuple:
            Command | None - command
            any - message data
        """

        def __disconnect(self):
            self._close_connection_processes()
            return False, None

        allowed_commands_id = allowed_commands.keys()

        # region LOGIC

        is_connected, received_message_list = self._receive_message_list()
        if received_message_list is None:
            return is_connected, None

        message_result_list = []

        for received_message_info in received_message_list:
            received_message = received_message_info

            received_command_id = received_message.command_id
            received_command = CCommandTypeEnum.get_command_by_id(received_command_id)

            # check if statemachine can fire
            if not GAME_STATE_MACHINE.can_fire(received_command.trigger.id):
                raise ValueError("Invalid state machine transition.")

            # Recieved Other
            if received_command_id not in allowed_commands_id:
                return __disconnect(self)

            is_processed = False
            for command_id, process_function in allowed_commands.items():
                if received_command_id == command_id:
                    # process_function(received_command, received_message)
                    command, message_data = process_function(received_command, received_message)
                    if not command:
                        is_connected = False
                        return is_connected, None

                    message_result_list.append((command, message_data))
                    is_processed = True
                    break

            if is_processed:
                continue

            raise ValueError("Invalid command ID for running game messages.")

        return is_connected, message_result_list

    def _process_respond_successs(self, received_command: Command) -> bool:
        is_connected = self._respond_client_success()
        GAME_STATE_MACHINE.send_trigger(received_command.trigger)
        return is_connected

    def _process_server_update_list(self, received_command, received_message) -> tuple[Command | None, list | None]:
        is_connected = self._process_respond_successs(received_command)
        if not is_connected:
            return None, None
        return received_command, received_message.get_array_param()

    def _respond_client_success(self) -> bool:
        command = CCommandTypeEnum.ResponseClientSuccess.value
        return self._send_message(command)
    # endregion

    # def receive_server_game_list_update(self) -> tuple[bool, GameList | None]:
    #     command = CCommandTypeEnum.ServerUpdateGameList.value
    #     return self._receive_standard_update_command(command)

    def receive_server_running_game_messages(self) -> tuple[bool, list[tuple[Command | None, GameData | None | str]]]:
        """

        :return:
        bool - is connected

        """

        def __process_standard(received_command, received_message) -> tuple[Command | None, None]:
            is_connected = self._process_respond_successs(received_command)
            if not is_connected:
                return None, None
            return received_command, None

        def __process_server_update_end_score(received_command, received_message) -> tuple[Command | None, str | None]:
            is_connected = self._process_respond_successs(received_command)
            if not is_connected:
                return None, None
            return received_command, received_message.get_single_param()

        allowed_commands = {
            CCommandTypeEnum.ServerUpdateGameData.value.id: self._process_server_update_list,
            CCommandTypeEnum.ServerStartTurn.value.id: __process_standard,
            CCommandTypeEnum.ServerUpdateEndScore.value.id: __process_server_update_end_score,
            CCommandTypeEnum.ServerUpdateNotEnoughPlayers.value.id: __process_standard,
            CCommandTypeEnum.ServerPingPlayer.value.id: self._process_server_ping_player
        }

        return self._receive_standard_state_messages(allowed_commands)

    def receive_lobby_messages(self) -> tuple[bool, list[tuple[Command | None, PlayerList | None]]]:
        """
        :return:
        bool - is timeout
        list of tuple:
            bool - is_connected
            Command | None - command
            PlayerList | None - message data
        """

        allowed_commands = {
            CCommandTypeEnum.ServerUpdateGameList.value.id: self._process_server_update_list,
            CCommandTypeEnum.ServerPingPlayer.value.id: self._process_server_ping_player
        }

        return self._receive_standard_state_messages(allowed_commands)

    def receive_before_game_messages(self) -> tuple[bool, list[tuple[Command | None, PlayerList | None]]]:
        """

        :return:
        bool - is timeout
        list of tuple:
            bool - is_connected
            Command | None - command
            PlayerList | None - message data
        """

        def __process_server_update_start_game(received_command, received_message) -> tuple[
            Command | None, PlayerList | None]:
            """

            :param received_command:
            :param received_message:
            :return:
                bool - is_connected
                Command | None - command
                PlayerList | None - message data
            """
            is_connected = self._process_respond_successs(received_command)
            if not is_connected:
                return None, None
            return received_command, None

        allowed_commands = {
            CCommandTypeEnum.ServerUpdatePlayerList.value.id: self._process_server_update_list,
            CCommandTypeEnum.ServerUpdateStartGame.value.id: __process_server_update_start_game,
            CCommandTypeEnum.ServerPingPlayer.value.id: self._process_server_ping_player
        }

        return self._receive_standard_state_messages(allowed_commands)

    def receive_server_game_data_messages(self) -> tuple[bool, list[tuple[Command | None, GameData | None]]]:
        """

        :return:
        """

        allowed_commands = {
            CCommandTypeEnum.ServerUpdateGameData.value.id: self._process_server_update_list,
            CCommandTypeEnum.ServerPingPlayer.value.id: self._process_server_ping_player
        }

        return self._receive_standard_state_messages(allowed_commands)

    # endregion

    # endregion

    # get nickname
    @property
    def nickname(self):
        return self._nickname

    # get was connected
    @property
    def was_connected(self):
        return self._was_connected

    @property
    def is_connected(self):
        return self._s is not None

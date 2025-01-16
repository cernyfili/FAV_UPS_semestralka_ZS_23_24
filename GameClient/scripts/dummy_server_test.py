import logging
import queue
import socket
import threading
import time

from src.backend.parser import convert_message_to_network_string
from src.frontend.ui_manager import MyApp
from src.shared.constants import CMessageConfig, CCommandTypeEnum, NetworkMessage, Param

WAIT_TIME = 0.5

response_queue = queue.Queue()

nickname = "Player1"


def create_response_dice_success():
    response = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerDiceSuccess.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return response


def simulate_server_response():
    def handle_client_connection(client_socket):
        # savet to file logging
        # logging.basicConfig(filename='dummy_server_server.log', level=logging.DEBUG)
        try:
            while True:
                # Simulate receiving a message from the client
                request = client_socket.recv(1024)
                if not request:
                    break

                request_id = request.decode()[6:8]
                command_name = CCommandTypeEnum.get_command_name_from_id(int(request_id))
                logging.debug(f"SERVER: Received request {command_name}: {request.decode()}")
                print(f"SERVER: Received request {command_name}: {request.decode()}")


                request_message = request.decode()

                client_login = "KIVUPS" + f"{CCommandTypeEnum.ClientLogin.value.id:02}"
                client_logout = "KIVUPS" + f"{CCommandTypeEnum.ClientLogout.value.id:02}"
                client_join_game = "KIVUPS" + f"{CCommandTypeEnum.ClientJoinGame.value.id:02}"
                client_start_game = "KIVUPS" + f"{CCommandTypeEnum.ClientStartGame.value.id:02}"
                client_roll_dice = "KIVUPS" + f"{CCommandTypeEnum.ClientRollDice.value.id:02}"
                client_selected_cubes = "KIVUPS" + f"{CCommandTypeEnum.ClientSelectedCubes.value.id:02}"

                if client_login in request_message:
                    message = create_response_game_list()
                    send_message(client_socket, message)

                    # ServerUpdateGameList
                    time.sleep(WAIT_TIME)
                    message = create_game_list_update()
                    send_message(client_socket, message)

                    # ServerUpdateGameList
                    time.sleep(WAIT_TIME)
                    message = create_game_list_update()
                    send_message(client_socket, message)

                elif client_logout in request_message:
                    response = create_response_success()
                    send_message(client_socket, response)

                elif client_join_game in request_message:
                    response = create_response_success()
                    send_message(client_socket, response)

                    # ServerUpdatePlayerList
                    time.sleep(WAIT_TIME)
                    message = create_player_list()
                    send_message(client_socket, message)

                elif client_start_game in request_message:
                    # ResponseServerSuccess
                    response = create_response_success()
                    send_message(client_socket, response)

                    # wait
                    time.sleep(WAIT_TIME)

                    # ServerUpdateGameData
                    message = create_game_data_update()
                    send_message(client_socket, message)

                    # wait
                    time.sleep(WAIT_TIME)

                    # ServerStartTurn
                    message = create_server_start_turn()
                    send_message(client_socket, message)

                    # ServerPingPlayer
                    time.sleep(WAIT_TIME)
                    message = create_server_ping_player()
                    send_message(client_socket, message)

                    # ServerUpdateGameData
                    time.sleep(WAIT_TIME)
                    message = create_game_data_update()
                    send_message(client_socket, message)

                elif client_roll_dice in request_message:
                    # ResponseServerSelectedCubes
                    response = create_response_server_select_cubes()
                    send_message(client_socket, response)

                    # ServerPingPlayer
                    time.sleep(WAIT_TIME)
                    message = create_server_ping_player()
                    send_message(client_socket, message)

                    # ServerUpdateGameData
                    time.sleep(WAIT_TIME)
                    message = create_game_data_update()
                    send_message(client_socket, message)

                elif client_selected_cubes in request_message:
                    # ResponseServerEndScore
                    # response = create_response_server_end_score()
                    # send_message(client_socket, response)

                    # ServerResponseDiceSuccess
                    response = create_response_dice_success()
                    send_message(client_socket, response)

                    # ServerPingPlayer
                    time.sleep(5)
                    response = create_server_ping_player()
                    send_message(client_socket, response)

                    # ServerPingPlayer
                    time.sleep(5)
                    response = create_server_ping_player()
                    send_message(client_socket, response)


                else:
                    # Get the response from the queue
                    if not response_queue.empty():
                        response = response_queue.get()
                        send_message(client_socket, response)
                    # Periodically send updates to the client
                    # update_message = create_game_list_update_response()
                    # print(f"Sending update: {update_message}")
                    # client_socket.sendall(update_message.encode())
                    # time.sleep(2)  # Send update every 5 seconds


        finally:
            client_socket.close()

    def send_message(client_socket, message):
        message_str = convert_message_to_network_string(message)
        logging.debug(f"SERVER: Sending response {CCommandTypeEnum.get_command_name_from_id(message.command_id)}:  {message}")
        print(f"SERVER: Sending response {CCommandTypeEnum.get_command_name_from_id(message.command_id)}:  {message}")
        print(
            f"SERVER: Sending response {CCommandTypeEnum.get_command_name_from_id(message.command_id)}:  {message_str}")
        client_socket.sendall(message_str.encode())

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(('localhost', 10000))
    server.listen(1)
    logging.debug("SERVER: Dummy server listening on port 10001")
    print("SERVER: Dummy server listening on port 10001")

    while True:
        client_socket, addr = server.accept()
        client_handler = threading.Thread(target=handle_client_connection, args=(client_socket,))
        client_handler.start()

def add_response_to_queue(response):
    response_queue.put(response)

switcher = True


def create_server_start_turn():
    server_start_turn = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerStartTurn.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return server_start_turn

def create_game_list_update():
    global switcher
    switcher = not switcher

    game_list_1 = "[{\"gameName\":\"GameA1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"};{\"gameName\":\"GameA2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]"
    game_list_2 = "[{\"gameName\":\"GameB1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"};{\"gameName\":\"GameB2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]"

    param_value = game_list_1 if switcher else game_list_2


    game_list_update = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerUpdateGameList.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("gameList", param_value)]
    )
    return game_list_update

def create_response_game_list():
    param_value = "[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]"
    game_list = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerGameList.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("gameList", param_value)]
    )
    return game_list

# main
def create_response_success():
    success = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerSuccess.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return success

def create_player_list():
    player_list = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerUpdatePlayerList.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("playerList", "[{\"playerName\":\"Player1\",\"isConnected\":\"1\"};{\"playerName\":\"Player2\",\"isConnected\":\"1\"};{\"playerName\":\"Player3\",\"isConnected\":\"0\"}]")]
    )
    return player_list

def create_game_data_update():
    game_data_update = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerUpdateGameData.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("gameData", "[{\"playerName\":\"Player1\",\"isConnected\":\"0\",\"score\":\"0\",\"isTurn\":\"0\"};{\"playerName\":\"Player2\",\"isConnected\":\"1\",\"score\":\"0\",\"isTurn\":\"1\"}]")]
    )
    return game_data_update

def create_response_server_select_cubes():
    response = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerSelectCubes.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("cubeValues", "[{\"value\":\"1\"};{\"value\":\"2\"};{\"value\":\"5\"};{\"value\":\"2\"};{\"value\":\"6\"};{\"value\":\"5\"}]")]
    )
    return response


def create_response_server_end_score():
    response = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerEndScore.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return response

def create_server_ping_player():
    response = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerPingPlayer.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return response


class CustomFormatter(logging.Formatter):
    def format(self, record):
        record.msg = record.msg.replace(':', '-')
        return super().format(record)


def main():

    app = MyApp()
    app.geometry("800x400")
    app.show_page("StartPage")
    app.mainloop()


if __name__ == "__main__":
    # Start the dummy server
    server_thread = threading.Thread(target=simulate_server_response, daemon=True)
    server_thread.start()

    logging.basicConfig(
        level=logging.DEBUG
    )

    for handler in logging.getLogger().handlers:
        handler.setFormatter(
            CustomFormatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s\n    ["%(filename)s:%(lineno)d"]'))
    main()

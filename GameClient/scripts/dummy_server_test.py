import logging
import socket
import threading
import queue
import time

from frontend.ui_manager import MyApp
from shared.constants import CMessageConfig, CCommandTypeEnum, NetworkMessage, Param
from backend.parser import convert_message_to_network_string, parse_message

response_queue = queue.Queue()

nickname = "Player1"

def simulate_server_response():
    def handle_client_connection(client_socket):
        try:
            while True:
                # Simulate receiving a message from the client
                request = client_socket.recv(1024)
                if not request:
                    break
                print(f"Received request: {request.decode()}")

                request_message = request.decode()
                substring = "KIVUPS" + f"{CCommandTypeEnum.ClientJoinGame.value.id:02}"
                if substring in request_message:
                    response = create_response_success()
                    print(f"Sending response: {response}")
                    client_socket.sendall(response.encode())
                else:
                    # Get the response from the queue
                    if not response_queue.empty():
                        response = response_queue.get()
                        print(f"Sending response: {response}")
                        client_socket.sendall(response.encode())
                    # Periodically send updates to the client
                    update_message = create_game_list_update_response()
                    print(f"Sending update: {update_message}")
                    client_socket.sendall(update_message.encode())
                    time.sleep(2)  # Send update every 5 seconds


        finally:
            client_socket.close()

    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(('localhost', 10000))
    server.listen(1)
    print("Dummy server listening on port 10001")

    while True:
        client_socket, addr = server.accept()
        client_handler = threading.Thread(target=handle_client_connection, args=(client_socket,))
        client_handler.start()

def add_response_to_queue(response):
    response_queue.put(response)

switcher = True

def create_game_list_update_response():
    global switcher
    switcher = not switcher

    param_value = "[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]" if switcher else "[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"};{\"gameName\":\"Game2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]"


    game_list_update = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ServerUpdateGameList.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("gameList", param_value)]
    )
    return convert_message_to_network_string(game_list_update)

def create_response_game_list():
    param_value = "[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]" if switcher else "[{\"gameName\":\"Game1\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"};{\"gameName\":\"Game2\",\"maxPlayers\":\"4\",\"connectedPlayers\":\"2\"}]"
    game_list = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerGameList.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[Param("gameList", param_value)]
    )
    return convert_message_to_network_string(game_list)

# main
def create_response_success():
    success = NetworkMessage(
        signature=CMessageConfig.SIGNATURE,
        command_id=CCommandTypeEnum.ResponseServerSuccess.value.id,
        timestamp="2025-01-05 12:00:00.000000",
        player_nickname=nickname,
        parameters=[]
    )
    return convert_message_to_network_string(success)


if __name__ == "__main__":
    logging.basicConfig(level=logging.DEBUG)

    # Start the dummy server
    server_thread = threading.Thread(target=simulate_server_response, daemon=True)
    server_thread.start()

    # Add a response to the queue
    response = create_response_game_list()
    add_response_to_queue(response)

    # Start the GUI application
    app = MyApp()
    app.geometry("800x400")
    app.show_page("StartPage")
    app.mainloop()
import random
import tkinter as tk
from datetime import datetime
from tkinter import messagebox
from typing import List

from constants import CommandTypeEnum, CGMessageSignature
from parser import NetworkMessage, convert_message_to_network_string, parse_message, Param, convert_params_game_list
from state_machine import G_game_state_machine


class MyApp(tk.Tk):
    def __init__(self):
        tk.Tk.__init__(self)
        self.title("Multi-Page Tkinter App")

        # Create container to hold pages
        self.container = tk.Frame(self)
        self.container.pack(side="top", fill="both", expand=True)
        self.container.grid_rowconfigure(0, weight=1)
        self.container.grid_columnconfigure(0, weight=1)

        # Dictionary to hold pages
        self.pages = {}

        # Create pages
        for PageClass in (StartPage, GameListPage, PlayerListPage, RunningGamePage, MyTurn):
            page_name = PageClass.__name__
            frame = PageClass(parent=self.container, controller=self)
            self.pages[page_name] = frame
            frame.grid(row=0, column=0, sticky="nsew")

        # Show the initial page
        self.show_page("StartPage")

    def show_page(self, page_name):
        # Show the selected page
        frame = self.pages[page_name]
        frame.tkraise()


import re
import socket


def connect_to_server(ip, port) -> socket.socket:
    # Create a socket object
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    # Connect to the server
    s.connect((ip, port))

    return s


def create_message(command_id, nickname, param):
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    message = NetworkMessage(CGMessageSignature, command_id, timestamp, nickname, param)
    return convert_message_to_network_string(message)


def convert_params_error_message(parameters: List[Param]) -> str:
    # Convert the parameters to a string
    if len(parameters) != 1:
        error_message = ""
    error_message = parameters[0].value

    return error_message

class Game:
    def __init__(self, game_name, connected_count_players, max_players):
        self.game_name = game_name
        self.connected_count_players = connected_count_players
        self.max_players = max_players



def connect_button_action(ip, port, nickname):

    try:
        connection = connect_to_server(ip, port)
    except all:
        # Show an error message if the connection fails
        messagebox.showerror("Connection Failed", "Unable to connect to the server!")
        return

    message_str = create_message(CommandTypeEnum.ClientLogin.value.id, nickname, [])
    # Send the nickname to the server
    try:
        connection.sendall(message_str.encode())
    except all:
        # Show an error message if the nickname sending fails
        messagebox.showerror("Connection Failed", "Unable to connect to the server!")
        return

    response_message_str = connection.recv(1024).decode()
    response_message = parse_message(response_message_str)

    if response_message.command_id == CommandTypeEnum.ResponseServerErrDuplicitNickname.value.id:
        # Show an error message if the nickname is already taken
        messagebox.showerror("Connection Failed", "Nickname is already taken!")
        return

    if response_message.command_id == CommandTypeEnum.ResponseServerError.value.id:
        error_message = convert_params_error_message(response_message.parameters)
        # Show an error message if the server responds with an error
        messagebox.showerror("Connection Failed", error_message)
        return

    if response_message.command_id != CommandTypeEnum.ResponseServerGameList.value.id:
        # Show an error message if the server responds with an unknown command
        messagebox.showerror("Connection Failed", "Unknown command received!")
        return

    game_list = convert_params_game_list(response_message.parameters)

    # State machine transition
    G_game_state_machine.send(CommandTypeEnum.ClientLogin.value.trigger)

    # Show the GameListPage
    app.show_page("GameListPage")


class StartPage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        label = tk.Label(self, text="Connect to a Server")
        label.pack(pady=10, padx=10, fill='both', expand=True)

        # Form with IP and Port and NickName
        ip_label = tk.Label(self, text="IP Address")
        ip_label.pack(pady=10, padx=10, fill='both', expand=True)
        ip_entry = tk.Entry(self, validate="focusout", validatecommand=(self.register(self.validate_ip), '%P'))
        ip_entry.pack(pady=10, padx=10, fill='both', expand=True)

        port_label = tk.Label(self, text="Port")
        port_label.pack(pady=10, padx=10, fill='both', expand=True)
        port_entry = tk.Entry(self, validate="focusout", validatecommand=(self.register(self.validate_port), '%P'))
        port_entry.pack(pady=10, padx=10, fill='both', expand=True)

        nickname_label = tk.Label(self, text="Nickname")
        nickname_label.pack(pady=10, padx=10, fill='both', expand=True)
        nickname_entry = tk.Entry(self, validate="focusout",
                                  validatecommand=(self.register(self.validate_nickname), '%P'))
        nickname_entry.pack(pady=10, padx=10, fill='both', expand=True)

        connect_button = tk.Button(self, text="Connect",
                                   command=lambda: connect_button_action(ip_entry.get(), port_entry.get(),
                                                                         nickname_entry.get()))
        connect_button.pack(pady=10, padx=10, fill='both', expand=True)

    def validate_ip(self, ip):
        # Check if ip is a valid IP address
        return bool(re.match(r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$', ip))

    def validate_port(self, port):
        # Check if port is a number and within the valid range
        return port.isdigit() and 1 <= int(port) <= 65535

    def validate_nickname(self, nickname):
        # Check if nickname is not empty
        return bool(nickname)


class GameListPage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        label = tk.Label(self, text="Available Games")
        label.pack(pady=10, padx=10)

        # Get the games information
        self.games = self.get_games_info()

        for game in self.games:
            game_name = game["game_name"]
            connected_count_players = game["connected_count_players"]
            max_players = game["max_players"]

            # Create a label for the game
            game_label = tk.Label(self, text=f"{game_name}: {connected_count_players}/{max_players}")
            game_label.pack(pady=10, padx=10)

            # Create a connect button for the game
            connect_button = tk.Button(self, text="Connect", command=lambda game=game: self.connect_to_game(game))
            connect_button.pack(pady=10, padx=10)

            # Disable the connect button if the game is full
            if connected_count_players >= max_players:
                connect_button.config(state="disabled")

    def set_games_info(self):


    def connect_to_game(self, game):
        self.controller.show_page("PlayerListPage")


class PlayerListPage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        # List of player names
        self.players = self.get_players_info()

        for player in self.players:
            # Create a label for each player
            player_label = tk.Label(self, text=player)
            player_label.pack(pady=10, padx=10)

        # Create the "Start Game" button
        start_game_button = tk.Button(self, text="Start Game", command=self.start_game)
        start_game_button.pack(pady=10, padx=10)

        # Disable the "Start Game" button if there are less than 2 players
        if len(self.players) < 2:
            start_game_button.config(state="disabled")

    def get_players_info(self):
        # Fetch the players information
        # This is a placeholder. Replace it with the actual logic to fetch the players information.
        players = ["Player 1", "Player 2"]
        return players

    def start_game(self):
        # Implement the logic to start the game
        self.controller.show_page("RunningGamePage")


class RunningGamePage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        # List of players and their scores
        self.players = self.get_players_info()

        for player in self.players:
            player_name = player["player_name"]
            score = player["score"]
            is_playing = player["is_playing"]

            # Create a label for each player and their score
            player_label = tk.Label(self, text=f"{player_name}: {score}")
            player_label.pack(pady=10, padx=10)

            # If it's the player's turn, display a waiting animation
            if is_playing:
                self.waiting_animation = tk.Label(self, text="Playing")
                self.waiting_animation.pack(pady=10, padx=10)
                self.animate()

        button = tk.Button(self, text="Go to MyTurn", command=lambda: controller.show_page("MyTurn"))
        button.pack(pady=10, padx=10)

    def get_players_info(self):
        # Fetch the players information
        # This is a placeholder. Replace it with the actual logic to fetch the players information.
        players = [
            {"player_name": "Player 1", "score": 10, "is_playing": True},
            {"player_name": "Player 2", "score": 15, "is_playing": False},
            # Add more players as needed
        ]
        return players

    def animate(self):
        # Get the current text of the waiting animation
        text = self.waiting_animation.cget("text")

        # Update the text of the waiting animation
        if text.count(".") < 3:
            text += "."
        else:
            text = "Playing"

        self.waiting_animation.config(text=text)

        # Schedule the next animation
        self.after(500, self.animate)


class MyTurn(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        # Create the "Roll Dice" button
        self.roll_dice_button = tk.Button(self, text="Roll Dice", command=self.roll_dice, state="normal")
        self.roll_dice_button.pack(pady=10, padx=10)

        # Create a list of Canvas widgets and Checkbutton widgets for the dice cubes
        self.dice_cube_vars = [tk.IntVar() for _ in range(6)]
        self.dice_cube_canvases = [tk.Canvas(self, width=40, height=40, bg="white") for _ in range(6)]
        self.dice_cube_checkbuttons = [tk.Checkbutton(self, text="", variable=self.dice_cube_vars[i]) for i in range(6)]
        for i in range(6):
            self.dice_cube_canvases[i].place(x=50 + i * 60, y=50)
            self.dice_cube_checkbuttons[i].place(x=50 + i * 60, y=100)

        # Create the "Send Selected Dice Cubes" button
        self.send_button = tk.Button(self, text="Send Selected Dice Cubes", command=self.send_selected_dice_cubes,
                                     state="normal")
        self.send_button.pack(pady=100, padx=10)

    def draw_dice_cube(self, canvas, x, y, dots):
        # Draw a red square
        canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")

        # Draw dots within the square
        dot_positions = [(x - 10, y - 10), (x, y - 10), (x + 10, y - 10),
                         (x - 10, y), (x, y), (x + 10, y)]

        for dot_x, dot_y in dot_positions[:dots]:
            canvas.create_oval(dot_x - 2, dot_y - 2, dot_x + 2, dot_y + 2, fill="white")

    def roll_dice(self):
        # Clear previous drawings on the canvases
        for canvas in self.dice_cube_canvases:
            canvas.delete("all")

        # Roll the dice and draw the dice cubes
        for i in range(6):
            dots = random.randint(1, 6)
            self.draw_dice_cube(self.dice_cube_canvases[i], 20, 20, dots)

            # Disable the checkbox for the non-valid cube
            if dots not in [1, 5]:
                self.dice_cube_checkbuttons[i].config(state="disabled")
            else:
                self.dice_cube_checkbuttons[i].config(state="normal")

        # Enable the "Send Selected Dice Cubes" button
        self.send_button.config(state="normal")

    def send_selected_dice_cubes(self):
        # Get the selected dice cubes
        selected_dice_cubes = [i for i, var in enumerate(self.dice_cube_vars) if var.get() == 1]

        # Implement the logic to send the selected dice cubes
        # This is a placeholder. Replace it with the actual logic to send the selected dice cubes.

        # Disable the "Send Selected Dice Cubes" button
        self.send_button.config(state="disabled")

        self.controller.show_page("RunningGamePage")


if __name__ == "__main__":
    app = MyApp()
    app.geometry("800x400")
    app.mainloop()

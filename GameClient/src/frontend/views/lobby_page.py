#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: lobby_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import tkinter as tk

from frontend.page_interface import PageInterface
from shared.constants import CGameConfig, CMessageConfig
from shared.data_structures import GameList


class LobbyPage(tk.Frame, PageInterface):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller
        self._game_list : GameList = GameList()

        self._load_page_content()

    def connect_to_game(self, game):
        self.controller.show_page("PlayerListPage")

    def _load_page_content(self):
        def show_game_list():
            game_list = self._game_list
            label = tk.Label(self, text="Available Games")
            label.pack(pady=10, padx=10)

            for game in game_list:
                game_name = game.name
                connected_count_players = game.connected_count_players
                max_players = game.max_players

                # Create a label for the game
                game_label = tk.Label(self, text=f"{game_name}: {connected_count_players}/{max_players}")
                game_label.pack(pady=10, padx=10)

                # Create a connect button for the game
                connect_button = tk.Button(self, text="Connect", command=lambda game_element=game: self.connect_to_game(game_element))
                connect_button.pack(pady=10, padx=10)

                # Disable the connect button if the game is full
                if connected_count_players >= max_players:
                    connect_button.config(state="disabled")

        def open_popup_new_game():
            popup = tk.Toplevel(self)
            popup.title("Create New Game")
            popup.geometry("300x200")

            # Game name label and entry
            game_name_label = tk.Label(popup, text="Game Name:")
            game_name_label.pack(pady=5)
            game_name_entry = tk.Entry(popup)
            game_name_entry.pack(pady=5)

            # Max players label and entry
            max_players_label = tk.Label(popup, text="Max Players:")
            max_players_label.pack(pady=5)
            max_players_entry = tk.Entry(popup)
            max_players_entry.pack(pady=5)

            # Validation function
            def validate_inputs():
                game_name = game_name_entry.get()
                try:
                    max_players = int(max_players_entry.get())
                except ValueError:
                    tk.messagebox.showerror("Invalid Input", "Max Players must be a number")
                    return False

                if not (CMessageConfig.is_valid_name(game_name)):
                    tk.messagebox.showerror("Invalid Input", f"Game Name must be between {CGameConfig.NAME_MIN_CHARS} and {CGameConfig.NAME_MAX_CHARS} characters")
                    return False

                if not (CGameConfig.MIN_PLAYERS <= max_players <= CGameConfig.MAX_PLAYERS):
                    tk.messagebox.showerror("Invalid Input", f"Max Players must be between {CGameConfig.MIN_PLAYERS} and {CGameConfig.MAX_PLAYERS}")
                    return False

                return True

            # Create game button
            create_button = tk.Button(popup, text="Create", command=lambda: validate_inputs() and self.create_game(game_name_entry.get(), max_players_entry.get()) and popup.destroy())
            create_button.pack(pady=10)

            # Close button
            close_button = tk.Button(popup, text="Close", command=popup.destroy)
            close_button.pack(pady=5)

        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        show_game_list()

        # button to create a new game
        create_game_button = tk.Button(self, text="Create New Game", command=lambda: open_popup_new_game())
        create_game_button.pack(pady=10, padx=10)



    def update_data(self, data : GameList):
        self._game_list = data
        self._load_page_content()

    def create_game(self, param, param1):
        print("Creating game")
        return True



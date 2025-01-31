#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: lobby_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import logging
import threading
import tkinter as tk
from abc import ABC
from tkinter import messagebox

from src.backend.server_communication import ServerCommunication
from src.frontend.page_interface import UpdateInterface
from src.frontend.views.utils import PAGES_DIC, list_start_listening_for_updates, destroy_elements
from src.shared.constants import CGameConfig, CMessageConfig, Game, Param, CCommandTypeEnum, MessageFormatError


class LobbyPage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        self._list = []
        self._lock = threading.Lock()  # Initialize the lock
        self._stop_event = threading.Event()
        self._update_thread = None

        self._load_page_content()

    def tkraise(self, aboveThis=None):
        page_name = self.winfo_name()
        logging.debug(f"Raising Page: {page_name}")
        # Call the original tkraise method
        super().tkraise(aboveThis)

        self._list = []
        self._lock = threading.Lock()  # Initialize the lock
        self._stop_event = threading.Event()
        self._update_thread = None

        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _get_state_name(self):
        return 'stateLobby'

    def _get_update_function(self):
        return ServerCommunication().receive_lobby_messages

    def _set_update_thread(self, param):
        self._update_thread = param

    def _load_page_content(self):
        def show_game_list():
            game_list = self._list
            label = tk.Label(self, text="Available Games")
            label.pack(pady=10, padx=10)

            if game_list == '[]' or not game_list:
                label = tk.Label(self, text="No games available")
                label.pack(pady=10, padx=10)
                return

            if not isinstance(game_list[0], Game):
                raise ValueError("Invalid game list format")

            for game in game_list:
                game : Game = game
                game_name = game.name
                connected_count_players = game.connected_count_players
                max_players = game.max_players

                # Create a label for the game
                game_label = tk.Label(self, text=f"{game_name}: {connected_count_players}/{max_players}")
                game_label.pack(pady=10, padx=10)

                # Create a connect button for the game
                connect_button = tk.Button(self, text="Connect",
                                           command=lambda game_element=game_name: self._button_action_connect_to_game(
                                               game_element))
                connect_button.pack(pady=10, padx=10)

                # Disable the connect button if the game is full
                if connected_count_players >= max_players:
                    connect_button.config(state="disabled")

        def open_popup_new_game():
            if hasattr(self, 'popup') and self.popup.winfo_exists():
                self.popup.lift()
                return

            popup = tk.Toplevel(self)
            popup.title("Create New Game")
            popup.geometry("300x200")

            # Game name label and entry
            default_game_name = "Game1"
            game_name_label = tk.Label(popup, text="Game Name:")
            game_name_label.pack(pady=5)
            game_name_entry = tk.Entry(popup)
            game_name_entry.insert(0, default_game_name)
            game_name_entry.pack(pady=5)

            # Max players label and entry
            max_players_label = tk.Label(popup, text="Max Players:")
            max_players_label.pack(pady=5)
            max_players_entry = tk.Entry(popup)
            max_players_entry = tk.Entry(popup)
            max_players_entry.insert(0, "4")
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
                    tk.messagebox.showerror("Invalid Input",
                                            f"Game Name must be between {CMessageConfig.NAME_MIN_CHARS} and {CMessageConfig.NAME_MAX_CHARS} characters")
                    return False

                if not (CGameConfig.MIN_PLAYERS <= max_players <= CGameConfig.MAX_PLAYERS):
                    tk.messagebox.showerror("Invalid Input", f"Max Players must be between {CGameConfig.MIN_PLAYERS} and {CGameConfig.MAX_PLAYERS}")
                    return False

                return True

            # Create game button
            create_button = tk.Button(popup, text="Create", command=lambda: validate_inputs() and self._button_action_create_game(game_name_entry.get(), int(max_players_entry.get())) and popup.destroy())
            create_button.pack(pady=10)

            # Close button
            close_button = tk.Button(popup, text="Close", command=popup.destroy)
            close_button.pack(pady=5)

        # Clear the current content except for popups
        destroy_elements(self)

        self._show_logout_button(tk)

        show_game_list()

        # button to create a new game
        create_game_button = tk.Button(self, text="Create New Game", command=lambda: open_popup_new_game())
        create_game_button.pack(pady=10, padx=10)

        self.bind('<Return>', lambda event: open_popup_new_game())

    def _button_action_create_game(self, game_name : str, max_players_count : int) -> bool:
        send_function = ServerCommunication().send_client_create_game
        next_page_name = PAGES_DIC.BeforeGamePage
        param_list = [Param("gameName", game_name), Param("maxPlayers", max_players_count)]

        return self.button_action_standard(tk=tk, send_function=send_function, next_page_name=next_page_name,
                                           param_list=param_list)

    def _button_action_connect_to_game(self, game_name: str) -> bool:
        send_function = ServerCommunication().send_client_join_game
        next_page_name = PAGES_DIC.BeforeGamePage
        param_list = [Param("gameName", game_name)]

        return self.button_action_standard(tk=tk, send_function=send_function, next_page_name=next_page_name,
                                           param_list=param_list)

    def _process_is_not_connected(self):

        next_page_name = PAGES_DIC.StartPage

        self.controller.show_page(next_page_name)

    def _start_listening_for_updates(self):
        process_command: dict[int, callable] = {
            CCommandTypeEnum.ResponseServerError.value.id: self.process_error
        }
        continue_commands = [CCommandTypeEnum.ServerPingPlayer.value]
        update_command = CCommandTypeEnum.ServerUpdateGameList.value

        #list_start_listening_for_updates(self, process_command, update_command, continue_commands, self._process_is_not_connected)
        list_start_listening_for_updates(self, process_command, update_command, continue_commands)

        # def on_thread_finish():
        #     print("Thread")
        #     with self._lock:
        #         if self._is_connected == False:
        #             next_page_name, page_data = process_is_not_connected()
        #             self.controller.show_page(next_page_name, page_data)
        #
        # threading.Thread(target=lambda: (self._update_thread.join(), on_thread_finish())).start()

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

from pyparsing import empty

from backend.server_communication import ServerCommunication
from frontend.page_interface import PageInterface, UpdateInterface
from frontend.views.utils import PAGES_DIC
from frontend.views.before_game_page import BeforeGamePage
from frontend.views.utils import process_is_not_connected, stop_update_thread
from shared.constants import CGameConfig, CMessageConfig, GameList, Game, GAME_STATE_MACHINE


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
        logging.debug("Raising Page")
        # Call the original tkraise method
        super().tkraise(aboveThis)
        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _get_state_name(self):
        return 'stateLobby'

    def _get_update_function(self):
        return ServerCommunication().receive_game_list_update

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
                connect_button = tk.Button(self, text="Connect", command=lambda game_element=game: self._button_action_connect_to_game(game_element))
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
                    tk.messagebox.showerror("Invalid Input", f"Game Name must be between {CGameConfig.NAME_MIN_CHARS} and {CGameConfig.NAME_MAX_CHARS} characters")
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
        for widget in self.winfo_children():
            if isinstance(widget, tk.Toplevel):
                continue
            widget.destroy()

        self._show_logout_button(tk)

        show_game_list()

        # button to create a new game
        create_game_button = tk.Button(self, text="Create New Game", command=lambda: open_popup_new_game())
        create_game_button.pack(pady=10, padx=10)

    def _button_action_create_game(self, game_name : str, max_players_count : int) -> bool:
        send_function = ServerCommunication().send_client_create_game
        next_page_name = PAGES_DIC.BeforeGamePage
        
        stop_update_thread(self)
        try:
            is_connected = send_function(game_name, max_players_count)
            if not is_connected:
                process_is_not_connected(self)
        except Exception as e:
            #messagebox.showerror("Connection Failed", str(e))
            #todo

            raise e
            return False

        self.controller.show_page(next_page_name)
        return True

    def _button_action_connect_to_game(self, game_name: str) -> bool:
        send_function = ServerCommunication().send_client_join_game
        next_page_name = PAGES_DIC.BeforeGamePage

        stop_update_thread(self)
        try:
            is_connected = send_function(game_name)
            if not is_connected:
                process_is_not_connected(self)
        except Exception as e:
            #messagebox.showerror("Connection Failed", str(e))
            #todo

            raise e
            return False

        self.controller.show_page(next_page_name)
        return True



#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: BeforeGamePage.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import logging
import threading
import tkinter as tk
from abc import ABC

from backend.server_communication import ServerCommunication
from frontend.page_interface import UpdateInterface
from frontend.views.utils import PAGES_DIC
from shared.constants import CGameConfig, PlayerList


class BeforeGamePage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller
        # List of player names
        self._list: PlayerList = PlayerList([])
        self._lock = threading.Lock()
        self._stop_event = threading.Event()

        self._load_page_content()

    def _get_state_name(self):
        return 'stateGame'

    def _get_update_function(self):
        return ServerCommunication().receive_server_player_list_update

    def _set_update_thread(self, param):
        self._update_thread = param

    def tkraise(self, aboveThis=None):
        page_name = self.winfo_name()
        logging.debug(f"Raising Page: {page_name}")
        # Call the original tkraise method
        super().tkraise(aboveThis)
        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _load_page_content(self):
        def show_players_list():
            player_list = self._list
            if player_list == '[]' or not player_list:
                label = tk.Label(self, text="No players connected")
                label.pack(pady=10, padx=10)
                return

            for player in self._list:
                # Create a label for each player
                player_label = tk.Label(self, text=player.name)
                player_label.pack(pady=10, padx=10)

                # set grey color if not connected
                if not player.is_connected:
                    player_label.config(fg="grey")

        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        self._show_logout_button(tk)

        show_players_list()

        # Create the "Start Game" button
        start_game_button = tk.Button(self, text="Start Game", command=self._button_action_start_game)
        start_game_button.pack(pady=10, padx=10)

        # Disable the "Start Game" button if there are less than 2 players connected
        connected_players_num = len([player for player in self._list if player.is_connected])
        if connected_players_num < CGameConfig.MIN_PLAYERS:
            start_game_button.config(state="disabled")

        self.bind('<Return>', lambda event: self._button_action_start_game())

    def _button_action_start_game(self) -> bool:

        return self.button_action_standard(tk=tk, send_function=ServerCommunication().send_client_start_game,
                                           next_page_name=PAGES_DIC.RunningGamePage, param_list=[])

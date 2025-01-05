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
from frontend.page_interface import PageInterface, UpdateInterface
from frontend.views.utils import process_is_not_connected, stop_update_thread
from shared.constants import CGameConfig, PlayerList, GAME_STATE_MACHINE


class BeforeGamePage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller
        # List of player names
        self._list : PlayerList = []
        self._lock = threading.Lock()
        self._stop_event = threading.Event()

        self._load_page_content()

    def _get_state_name(self):
        return 'BeforeGamePage'

    def _get_update_function(self):
        return ServerCommunication().receive_player_list_update

    def _set_update_thread(self, param):
        self._update_thread = param

    def tkraise(self, aboveThis=None):
        logging.debug("Raising Page")
        # Call the original tkraise method
        super().tkraise(aboveThis)
        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    # def _start_listening_for_updates(self):
    #     state_name = 'BeforeGamePage'
    #     update_function = ServerCommunication().receive_player_list_update
    #
    #     logging.debug("Starting to listen for updates")
    #     def listen_for_updates():
    #         current_state = GAME_STATE_MACHINE.get_current_state()
    #         while current_state == state_name and not self._stop_event.is_set():
    #             logging.debug("Listening for game list updates")
    #             try:
    #                 update_list = update_function()
    #                 if self._stop_event.is_set():
    #                     break
    #                 if not update_list:
    #                     break
    #                 self.update_data(update_list)
    #             except Exception as e:
    #                 raise e
    #                 break
    #     # Wait for the update thread to finish
    #     self.update_thread = threading.Thread(target=listen_for_updates, daemon=True)
    #     self.update_thread.start()

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

        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        show_players_list()

        # Create the "Start Game" button
        start_game_button = tk.Button(self, text="Start Game", command=self._button_action_start_game)
        start_game_button.pack(pady=10, padx=10)

        # Disable the "Start Game" button if there are less than 2 players
        if len(self._list) < CGameConfig.MIN_PLAYERS:
            start_game_button.config(state="disabled")

    def _button_action_start_game(self) -> bool:
        send_function = ServerCommunication().send_start_game
        next_page_name = "RunningGamePage"

        stop_update_thread(self)

        try:
            is_connected = send_function()
            if not is_connected:
                process_is_not_connected(self)
        except Exception as e:
            #messagebox.showerror("Connection Failed", str(e))
            #todo

            raise e
            return False

        self.controller.show_page(next_page_name)
        return True

    # def update_data(self, data : PlayerList):
    #     with self._lock:
    #         self._list = data
    #         self._load_page_content()
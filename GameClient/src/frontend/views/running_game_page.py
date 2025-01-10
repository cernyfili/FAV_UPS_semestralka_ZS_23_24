#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: running_game_page.py
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

from backend.server_communication import ServerCommunication
from frontend.page_interface import UpdateInterface
from frontend.views.utils import PAGES_DIC, show_game_data, game_data_start_listening_for_updates
from frontend.views.utils import process_is_not_connected
from shared.constants import GameData, GAME_STATE_MACHINE, CCommandTypeEnum


class RunningGamePage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        # List of players and their scores
        self._list : GameData = GameData([])
        self._lock = threading.Lock()
        self._stop_event = threading.Event()

        self._load_page_content()

    def _get_state_name(self):
        return 'stateRunningGame'

    def _get_update_function(self):
        return ServerCommunication().receive_running_game_messages

    def _set_update_thread(self, param):
        self._update_thread = param

    def tkraise(self, aboveThis=None):
        logging.debug("Raising Page")
        # Call the original tkraise method
        super().tkraise(aboveThis)
        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _start_listening_for_updates(self):
        process_command = {
            CCommandTypeEnum.ServerStartTurn.value: self._process_start_turn,
            CCommandTypeEnum.ServerUpdateEndScore.value: self._process_update_end_score
        }

        game_data_start_listening_for_updates(self, process_command)

    def animate(self, waiting_animation):
        # Check if the widget still exists
        if not waiting_animation.winfo_exists():
            return

        # Get the current text of the waiting animation
        text = waiting_animation.cget("text")

        # Update the text of the waiting animation
        if text.count(".") < 3:
            text += "."
        else:
            text = "Playing"

        waiting_animation.config(text=text)

        # Schedule the next animation
        self.after(500, lambda: self.animate(waiting_animation))

    def _load_page_content(self):
        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        self._show_logout_button(tk)
        # header
        header = tk.Label(self, text="Game is running")
        header.pack(pady=10, padx=10)

        show_game_data(self, tk, self._list)

    def _process_start_turn(self):
        next_page_name = PAGES_DIC.MyTurnRollDicePage

        self.controller.show_page(next_page_name)

    def _process_update_end_score(self, player_name):
        next_page_name = PAGES_DIC.LobbyPage

        messagebox.showinfo("End of game", f"Player {player_name} has WON")
        self.controller.show_page(next_page_name)


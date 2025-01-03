#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: BeforeGamePage.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import tkinter as tk

from frontend.page_interface import PageInterface
from shared.constants import CGameConfig
from shared.data_structures import PlayerList


class BeforeGamePage(tk.Frame, PageInterface):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller


        # List of player names
        self.players : PlayerList = PlayerList()

        self._load_page_content()

    def _load_page_content(self):
        def show_players_list():
            for player in self.players:
                # Create a label for each player
                player_label = tk.Label(self, text=player.name)
                player_label.pack(pady=10, padx=10)

        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        show_players_list()

        # Create the "Start Game" button
        start_game_button = tk.Button(self, text="Start Game", command=self.start_game)
        start_game_button.pack(pady=10, padx=10)

        # Disable the "Start Game" button if there are less than 2 players
        if len(self.players) < CGameConfig.MIN_PLAYERS:
            start_game_button.config(state="disabled")

    def start_game(self):
        pass

    def update_data(self, data : PlayerList):
        self.players = data

        self._load_page_content()
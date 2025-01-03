#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: running_game_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import tkinter as tk

from frontend.page_interface import PageInterface
from shared.data_structures import GameData


class RunningGamePage(tk.Frame, PageInterface):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        # List of players and their scores
        self.gameData : GameData = GameData()

        self._load_page_content()

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

    def _load_page_content(self):
        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        # header
        header = tk.Label(self, text="Game is running")
        header.pack(pady=10, padx=10)



        for player in self.gameData:
            player_name = player.player_name
            score = player.score
            is_turn = player.is_turn
            is_connected = player.is_connected

            # Create a label for each player and their score
            player_label = tk.Label(self, text=f"{player_name}: {score}")


            # if the player isnt connected show the label in gray
            if not is_connected:
                player_label.config(fg="gray")

            player_label.pack(pady=10, padx=10)

            # If it's the player's turn, display a waiting animation
            if is_turn:
                self.waiting_animation = tk.Label(self, text="Playing")
                self.waiting_animation.pack(pady=2, padx=10)
                self.animate()

    def update_data(self, data : GameData):
        self.gameData = data

        self._load_page_content()

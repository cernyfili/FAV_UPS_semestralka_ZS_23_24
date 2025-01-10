#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: utils.py
Author: Filip Cerny
Created: 05.01.2025
Version: 1.0
Description: 
"""
import logging
import threading
import tkinter
from tkinter import messagebox
from types import SimpleNamespace

from shared.constants import reset_game_state_machine, GAME_STATE_MACHINE, CCommandTypeEnum

pages_names = {
    "StartPage": "StartPage",
    "LobbyPage": "LobbyPage",
    "BeforeGamePage": "BeforeGamePage",
    "RunningGamePage": "RunningGamePage",
    "MyTurnRollDicePage": "MyTurnRollDicePage",
    "MyTurnSelectCubesPage": "MyTurnSelectCubesPage",
}

PAGES_DIC = SimpleNamespace(**pages_names)

def process_is_not_connected(instance_object):
    # todo messagebox.showerror("Connection Failed", "Could not connect to the server")
    instance_object.controller.show_page("StartPage")
    raise ConnectionError("Could not connect to the server")

def stop_update_thread(self):
    # region THREAD STOP
    self._stop_event.set()
    # Wait for the update thread to finish
    if self._update_thread.is_alive():
        self._update_thread.join()
    # endregion


def show_game_data(self, tk, player_list):
    if player_list == '[]' or not player_list:
        label = tk.Label(self, text="No players connected")
        label.pack(pady=10, padx=10)
        return

    for player in player_list:
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
            self.animate(self.waiting_animation)


def game_data_start_listening_for_updates(self, process_command):
    state_name = self._get_state_name()
    update_function = self._get_update_function()


    logging.debug("Starting to listen for updates")
    def listen_for_updates():
        logging.debug("Listening for updates")
        current_state = GAME_STATE_MACHINE.get_current_state()
        while current_state == state_name and not self._stop_event.is_set():
            logging.debug("Listening for Data updates")
            try:
                is_connected, command, message_data = update_function()
                if self._stop_event.is_set():
                    break

                if not is_connected:
                    process_is_not_connected(self)
                    break

                # timeout trying read update message
                if not command and not message_data:
                    continue

                for command, handler in process_command.items():
                    if command == command:
                        handler()
                        break

                if command == CCommandTypeEnum.ServerUpdateGameData.value:
                    self.update_data(message_data)
                    continue

                raise Exception("Unknown command")
            except Exception as e:
                raise e
                # todo change
                messagebox.showerror("Error", str(e))
                break
    # Wait for the update thread to finish
    self._set_update_thread(threading.Thread(target=listen_for_updates, daemon=True))
    self._update_thread.start()

def my_turn_start_listening_for_updates(self):
    def _process_server_ping_player():
        pass

    process_command = {
        CCommandTypeEnum.ServerPingPlayer.value: _process_server_ping_player
    }

    game_data_start_listening_for_updates(self, process_command)

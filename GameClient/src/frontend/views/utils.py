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
from tkinter import messagebox
from types import SimpleNamespace

from src.backend.server_communication import ServerCommunication
from src.shared.constants import GAME_STATE_MACHINE, CCommandTypeEnum, Command

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
    logging.debug("Stopping the update thread")
    # region THREAD STOP
    self._stop_event.set()
    # Wait for the update thread to finish
    if self._update_thread.is_alive():
        logging.debug("Waiting for the update thread to finish")
        self._update_thread.join()
    # endregion


# region Loading animation
def stop_loading_animation(self):
    if hasattr(self, 'loading_label'):
        self.loading_label.destroy()


def bind_to_enter_key(self, button):
    pass
    # self.bind('<Return>', button.invoke())


def animate(self, waiting_animation, label_str):
    if not waiting_animation.winfo_exists():
        return

    text = waiting_animation.cget("text")
    if text.count(".") < 3:
        text += "."
    else:
        text = label_str

    waiting_animation.config(text=text)
    self.after(500, lambda: animate(self, waiting_animation, label_str))


def show_loading_animation(self, tk):
    # destroy the current content
    for widget in self.winfo_children():
        widget.destroy()

    self.loading_label = tk.Label(self, text="Loading...")
    self.loading_label.pack(pady=10, padx=10)
    animate(self, self.loading_label, "Loading")


# endregion

def show_game_data(self, tk, player_list):
    if player_list == '[]' or not player_list:
        label = tk.Label(self, text="No players connected")
        label.pack(pady=10, padx=10)
        return

    # Create a frame to contain the player data with a border
    frame = tk.Frame(self, bd=2, relief="solid")
    frame.pack(pady=10, padx=10)

    # Create a label for the game
    label = tk.Label(frame, text="GAME DATA", font=("Helvetica", 12, "underline"))
    label.pack(pady=10, padx=10)


    for player in player_list:
        player_name = player.player_name
        score = player.score
        is_turn = player.is_turn
        is_connected = player.is_connected

        if player_name == ServerCommunication().nickname:
            player_name += " (You)"

        # Create a label for each player and their score
        player_label = tk.Label(frame, text=f"{player_name}: {score}")


        # if the player isnt connected show the label in gray
        if not is_connected:
            player_label.config(fg="gray")

        player_label.pack(pady=10, padx=10)

        # If it's the player's turn, display a waiting animation
        if is_turn:
            self.waiting_animation = tk.Label(frame, text="Playing")
            self.waiting_animation.pack(pady=2, padx=10)
            animate(self=self, waiting_animation=self.waiting_animation, label_str="Playing")


def _standard_start_listening_for_updates(self, process_command, listen_for_updates: callable, continue_commands):
    pass


def list_start_listening_for_updates(self, process_command_dic: dict[int, callable], update_command: Command,
                                     continue_commands: list[Command]):
    def listen_for_updates(state_name: str, update_function: callable, process_command_dic: dict[int, callable],
                           continue_commands_list: list[Command]):
        logging.debug("Listening for updates")
        current_state = GAME_STATE_MACHINE.get_current_state()
        while current_state == state_name and not self._stop_event.is_set():
            # while not self._stop_event.is_set():
            logging.debug("Listening for Data updates")
            try:
                is_connected, message_info_list = update_function()
                if self._stop_event.is_set():
                    return

                if not is_connected:
                    process_is_not_connected(self)
                    return

                if not message_info_list:
                    # timeout trying read update message
                    continue

                for i, message_info in enumerate(message_info_list):
                    command, message_data = message_info

                    for command_id, handler in process_command_dic.items():
                        if command_id == command.id:
                            handler()
                            is_last_message = i == len(message_info_list) - 1
                            if is_last_message:
                                return

                    is_continue_command = False
                    for continue_command in continue_commands_list:
                        if command.id == continue_command.id:
                            is_continue_command = True
                    if is_continue_command:
                        continue

                    if command == update_command:
                        self.update_data(message_data)
                        continue

                    raise Exception("Unknown command")
            except Exception as e:
                raise e
                # todo change
                messagebox.showerror("Error", str(e))
                break

    state_name = self._get_state_name()
    update_function = self._get_update_function()

    logging.debug("Starting to listen for updates")
    # Wait for the update thread to finish
    self._set_update_thread(
        threading.Thread(target=listen_for_updates,
                         args=(state_name, update_function, process_command_dic, continue_commands), daemon=True))
    self._update_thread.start()

    # if command == CCommandTypeEnum.ServerPingPlayer.value:
    #     continue
    #
    # if command == CCommandTypeEnum.ServerUpdateGameData.value:
    #     self.update_data(message_data)
    #     continue

def my_turn_start_listening_for_updates(self):
    process_command = {
    }
    continue_commands = [CCommandTypeEnum.ServerPingPlayer.value]
    update_command = CCommandTypeEnum.ServerUpdateGameData.value
    list_start_listening_for_updates(self, process_command, update_command, continue_commands)

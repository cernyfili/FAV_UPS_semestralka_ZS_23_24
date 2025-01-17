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

from src.backend.server_communication import ServerCommunication
from src.frontend.page_interface import UpdateInterface
from src.frontend.views.utils import PAGES_DIC, show_game_data, process_is_not_connected
from src.shared.constants import GameData, CCommandTypeEnum, Command, GAME_STATE_MACHINE


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
        return ServerCommunication().receive_server_running_game_messages

    def _set_update_thread(self, param):
        self._update_thread = param

    def tkraise(self, aboveThis=None):
        page_name = self.winfo_name()
        logging.debug(f"Raising Page: {page_name}")
        # Call the original tkraise method
        super().tkraise(aboveThis)

        self._list : GameData = GameData([])
        self._lock = threading.Lock()
        self._stop_event = threading.Event()

        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _start_listening_for_updates(self):
        process_command: dict[int, callable] = {
            CCommandTypeEnum.ServerStartTurn.value.id: self._process_start_turn,
            #CCommandTypeEnum.ServerUpdateEndScore.value.id: self._process_update_end_score
        }
        update_command = CCommandTypeEnum.ServerUpdateGameData.value

        def listen_for_updates(state_name: str, update_function: callable, process_command_dic: dict[int, callable],
                               continue_commands_list: list[Command]):
            # logging.debug("Listening for updates")
            current_state = GAME_STATE_MACHINE.get_current_state()
            while current_state == state_name and not self._stop_event.is_set():
                # while not self._stop_event.is_set():
                # logging.debug("Listening for Data updates")
                try:
                    is_connected, message_info_list = update_function()
                    if self._stop_event.is_set():
                        page_name = self.winfo_name()
                        logging.debug(f"Stopped listening for updates on page: {page_name}")
                        return

                    if not is_connected:
                        process_is_not_connected(self)
                        page_name = self.winfo_name()
                        logging.debug(f"Stopped listening for updates on page: {page_name}")
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
                                    page_name = self.winfo_name()
                                    logging.debug(f"Stopped listening for updates on page: {page_name}")
                                    return

                        is_continue_command = False
                        for continue_command in continue_commands_list:
                            if command.id == continue_command.id:
                                is_continue_command = True
                        if is_continue_command:
                            continue
                        if command.id == CCommandTypeEnum.ServerUpdateEndScore.value.id:
                            self._process_update_end_score(message_data)
                            return

                        if command == update_command:
                            self.update_data(message_data)
                            continue

                        raise Exception("Unknown command")
                except Exception as e:
                    raise e
                    # todo change
                    messagebox.showerror("Error", str(e))
                    break
            page_name = self.winfo_name()
            logging.debug(f"Stopped listening for updates on page: {page_name}")


        state_name = self._get_state_name()
        update_function = self._get_update_function()

        page_name = self.winfo_name()
        logging.debug(f"Starting listening for updates on page: {page_name}")
        # Wait for the update thread to finish
        self._set_update_thread(
            threading.Thread(target=listen_for_updates,
                             args=(state_name, update_function, process_command, []), daemon=True))
        self._update_thread.start()

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


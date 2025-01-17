#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: my_turn_page.py
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
from src.frontend.views.utils import PAGES_DIC, show_game_data, my_turn_start_listening_for_updates, \
    show_loading_animation, \
    stop_loading_animation
from src.frontend.views.utils import process_is_not_connected, stop_update_thread
from src.shared.constants import CCommandTypeEnum, GameData


class MyTurnRollDicePage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        self._list : GameData = GameData([])
        self._lock = threading.Lock()  # Initialize the lock
        self._stop_event = threading.Event()
        self._update_thread = None

        self._load_page_content()

    def tkraise(self, aboveThis=None):
        page_name = self.winfo_name()
        logging.debug(f"Raising Page: {page_name}")
        # Call the original tkraise method
        super().tkraise(aboveThis)

        self._list : GameData = GameData([])
        self._lock = threading.Lock()  # Initialize the lock
        self._stop_event = threading.Event()
        self._update_thread = None

        # Custom behavior after raising the frame
        self._load_page_content()

        self._start_listening_for_updates()

    def _get_state_name(self):
        return 'stateMyTurn'

    def _get_update_function(self):
        return ServerCommunication().receive_server_my_turn_messages

    def _set_update_thread(self, param):
        self._update_thread = param

    def _load_page_content(self):
        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()

        self._show_logout_button(tk)

        # Create the "Roll Dice" button
        self.roll_dice_button = tk.Button(self, text="Roll Dice", command=self._button_action_send_roll_dice, state="normal")
        self.roll_dice_button.pack(pady=10, padx=10)

        show_game_data(self, tk, self._list)

    def _start_listening_for_updates(self):
        my_turn_start_listening_for_updates(self)

    def _button_action_send_roll_dice(self):
        def run_send_function():

            stop_update_thread(self)
            try:
                is_connected, command, cube_values_list = ServerCommunication().send_client_roll_dice()
                if not is_connected:
                    process_is_not_connected(self)

                if command == CCommandTypeEnum.ResponseServerSelectCubes.value:
                    next_page_name = PAGES_DIC.MyTurnSelectCubesPage
                elif command == CCommandTypeEnum.ResponseServerEndTurn.value:
                    messagebox.showinfo("End of turn",
                                        "Your turn is over you didnt roll any from the allowed combinations")
                    next_page_name = PAGES_DIC.RunningGamePage
                else:
                    raise Exception("Unknown command")
            except Exception as e:
                process_is_not_connected(self)
                # todo remove
                raise e
            finally:
                stop_loading_animation(self)

            self.controller.show_page(next_page_name, cube_values_list)

        show_loading_animation(self, tk)

        threading.Thread(target=run_send_function, args=()).start()
        return True

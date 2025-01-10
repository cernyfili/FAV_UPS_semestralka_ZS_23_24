#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: my_turn_select_cubes_page.py
Author: Filip Cerny
Created: 08.01.2025
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
from frontend.views.utils import PAGES_DIC, my_turn_start_listening_for_updates
from frontend.views.utils import process_is_not_connected, stop_update_thread
from shared.constants import CubeValuesList, ALLOWED_CUBE_VALUES_COMBINATIONS, CombinationList, CCommandTypeEnum, \
    GameData


class MyTurnSelectCubesPage(tk.Frame, UpdateInterface, ABC):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.cubes_values_list = None
        self.controller = controller

        self._list : GameData = GameData([])
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
        return 'stateNextDice'

    def _get_update_function(self):
        return ServerCommunication().receive_my_turn_messages

    def _set_update_thread(self, param):
        self._update_thread = param

    def _load_page_content(self):
        # Clear the current content
        for widget in self.winfo_children():
            widget.destroy()
        self._show_logout_button(tk)



        self._show_cube_values_list()

        # Create the "Send Selected Dice Cubes" button
        self.send_button = tk.Button(self, text="Send Selected Dice Cubes", command=self._button_action_send_selected_dice_cubes,
                                     state="normal")

        self.send_button.pack(side="bottom", pady=10, padx=10)

    def _start_listening_for_updates(self):
        my_turn_start_listening_for_updates(self)

    def update_data(self, gui_data):
        # called from MyApp.show_page(
        if isinstance(gui_data, CubeValuesList):
            self.cubes_values_list = gui_data
        # called from ServerCommunication.receive_my_turn_messages
        elif isinstance(gui_data, GameData):
            self._list = gui_data
        else:
            raise ValueError("Invalid data type")

        self._load_page_content()

    def _draw_dice_cube(self, canvas, x, y, dots):
        def draw_dice_cube_1(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 2, y - 2, x + 2, y + 2, fill="white")

        def draw_dice_cube_2(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 10, y - 10, x - 6, y - 6, fill="white")
            canvas.create_oval(x + 6, y + 6, x + 10, y + 10, fill="white")

        def draw_dice_cube_3(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 10, y - 10, x - 6, y - 6, fill="white")
            canvas.create_oval(x - 2, y - 2, x + 2, y + 2, fill="white")
            canvas.create_oval(x + 6, y + 6, x + 10, y + 10, fill="white")

        def draw_dice_cube_4(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 10, y - 10, x - 6, y - 6, fill="white")
            canvas.create_oval(x + 6, y - 10, x + 10, y - 6, fill="white")
            canvas.create_oval(x - 10, y + 6, x - 6, y + 10, fill="white")
            canvas.create_oval(x + 6, y + 6, x + 10, y + 10, fill="white")

        def draw_dice_cube_5(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 10, y - 10, x - 6, y - 6, fill="white")
            canvas.create_oval(x + 6, y - 10, x + 10, y - 6, fill="white")
            canvas.create_oval(x - 2, y - 2, x + 2, y + 2, fill="white")
            canvas.create_oval(x - 10, y + 6, x - 6, y + 10, fill="white")
            canvas.create_oval(x + 6, y + 6, x + 10, y + 10, fill="white")

        def draw_dice_cube_6(canvas, x, y):
            canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")
            canvas.create_oval(x - 10, y - 10, x - 6, y - 6, fill="white")
            canvas.create_oval(x + 6, y - 10, x + 10, y - 6, fill="white")
            canvas.create_oval(x - 10, y, x - 6, y + 4, fill="white")
            canvas.create_oval(x + 6, y, x + 10, y + 4, fill="white")
            canvas.create_oval(x - 10, y + 6, x - 6, y + 10, fill="white")
            canvas.create_oval(x + 6, y + 6, x + 10, y + 10, fill="white")

        draw_dice_cube_functions = [draw_dice_cube_1, draw_dice_cube_2, draw_dice_cube_3, draw_dice_cube_4, draw_dice_cube_5, draw_dice_cube_6]
        draw_dice_cube_functions[dots - 1](canvas, x, y)

    def _show_cube_values_list(self):

        cube_values_list = self.cubes_values_list
        if not cube_values_list:
            return

        cubes_num = len(self.cubes_values_list)

        # Create a list of Canvas widgets and Checkbutton widgets for the dice cubes
        self.dice_cube_vars = [tk.IntVar() for _ in range(cubes_num)]
        self.dice_cube_canvases = [tk.Canvas(self, width=40, height=40, bg="white") for _ in range(cubes_num)]
        self.dice_cube_checkbuttons = [tk.Checkbutton(self, text="", variable=self.dice_cube_vars[i]) for i in range(cubes_num)]
        for i in range(cubes_num):
            self.dice_cube_canvases[i].place(x=50 + i * 60, y=50)
            self.dice_cube_checkbuttons[i].place(x=50 + i * 60, y=100)



        allowed_list = ALLOWED_CUBE_VALUES_COMBINATIONS.create_allowed_values_mask(cube_values_list)

        # Roll the dice and draw the dice cubes
        for i, value in enumerate(cube_values_list):
            dots = value
            self._draw_dice_cube(self.dice_cube_canvases[i], 20, 20, dots)

            # Disable the checkbox for the non-valid cube
            if not allowed_list[i]:
                self.dice_cube_checkbuttons[i].config(state="disabled")

    def _button_action_send_selected_dice_cubes(self):
        def __is_valid_selected_dice_cubes(selected_dice_cubes):
            def ___remove_combination_from_list(cube_list, combination):
                for value in combination:
                    cube_list.remove(value)
                return cube_list
            # Check if the selected dice cubes are valid
            if not selected_dice_cubes:
                return False
            cube_list = selected_dice_cubes.copy()
            allowed_combinations = ALLOWED_CUBE_VALUES_COMBINATIONS.list
            changed_cube_list = cube_list.copy()
            while True:
                for combination in allowed_combinations:
                    if CombinationList.is_combination_in_list(combination=combination, cube_list=cube_list):
                        changed_cube_list = ___remove_combination_from_list(cube_list=changed_cube_list, combination=combination)
                if changed_cube_list == cube_list:
                    break
                cube_list = changed_cube_list.copy()

            return not cube_list
        # Get the selected dice cubes
        selected_dice_cubes_values = [self.cubes_values_list[i] for i in range(len(self.dice_cube_vars)) if self.dice_cube_vars[i].get()]

        if not __is_valid_selected_dice_cubes(selected_dice_cubes_values):
            rules_str = str(ALLOWED_CUBE_VALUES_COMBINATIONS)
            messagebox.showerror("Invalid Selection", "Please select valid dice cubes\n\nAllowed combinations:\n" + rules_str)
            return

        stop_update_thread(self)

        try:
            is_connected, command = ServerCommunication().send_client_select_cubes(selected_dice_cubes_values)
            if not is_connected:
                process_is_not_connected(self)

            if command == CCommandTypeEnum.ResponseServerDiceSuccess.value:
                next_page_name = PAGES_DIC.MyTurnRollDicePage
            elif command == CCommandTypeEnum.ResponseServerEndScore.value:
                messagebox.showinfo("!!! YOU WON !!!", "You have reached the score limit")
                next_page_name = PAGES_DIC.LobbyPage
            else:
                raise Exception("Unknown command")
        except Exception as e:
            #messagebox.showerror("Connection Failed", str(e))
            #todo

            raise e
            return

        self.controller.show_page(next_page_name)

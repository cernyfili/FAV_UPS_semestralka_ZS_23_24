#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: my_turn_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import random
import tkinter as tk

from frontend.page_interface import PageInterface


class MyTurnPage(tk.Frame, PageInterface):
    def _load_page_content(self):
        # Create the "Roll Dice" button
        self.roll_dice_button = tk.Button(self, text="Roll Dice", command=self.roll_dice, state="normal")
        self.roll_dice_button.pack(pady=10, padx=10)

        # Create a list of Canvas widgets and Checkbutton widgets for the dice cubes
        self.dice_cube_vars = [tk.IntVar() for _ in range(6)]
        self.dice_cube_canvases = [tk.Canvas(self, width=40, height=40, bg="white") for _ in range(6)]
        self.dice_cube_checkbuttons = [tk.Checkbutton(self, text="", variable=self.dice_cube_vars[i]) for i in range(6)]
        for i in range(6):
            self.dice_cube_canvases[i].place(x=50 + i * 60, y=50)
            self.dice_cube_checkbuttons[i].place(x=50 + i * 60, y=100)

        # Create the "Send Selected Dice Cubes" button
        self.send_button = tk.Button(self, text="Send Selected Dice Cubes", command=self.send_selected_dice_cubes,
                                     state="normal")
        self.send_button.pack(pady=100, padx=10)

    def update_data(self, data):
        pass

    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        self._load_page_content()



    def draw_dice_cube(self, canvas, x, y, dots):
        # Draw a red square
        canvas.create_rectangle(x - 20, y - 20, x + 20, y + 20, fill="red")

        # Draw dots within the square
        dot_positions = [(x - 10, y - 10), (x, y - 10), (x + 10, y - 10),
                         (x - 10, y), (x, y), (x + 10, y)]

        for dot_x, dot_y in dot_positions[:dots]:
            canvas.create_oval(dot_x - 2, dot_y - 2, dot_x + 2, dot_y + 2, fill="white")

    def roll_dice(self):
        # Clear previous drawings on the canvases
        for canvas in self.dice_cube_canvases:
            canvas.delete("all")

        # Roll the dice and draw the dice cubes
        for i in range(6):
            dots = random.randint(1, 6)
            self.draw_dice_cube(self.dice_cube_canvases[i], 20, 20, dots)

            # Disable the checkbox for the non-valid cube
            if dots not in [1, 5]:
                self.dice_cube_checkbuttons[i].config(state="disabled")
            else:
                self.dice_cube_checkbuttons[i].config(state="normal")

        # Enable the "Send Selected Dice Cubes" button
        self.send_button.config(state="normal")

    def send_selected_dice_cubes(self):
        # Get the selected dice cubes
        selected_dice_cubes = [i for i, var in enumerate(self.dice_cube_vars) if var.get() == 1]

        # Implement the logic to send the selected dice cubes
        # This is a placeholder. Replace it with the actual logic to send the selected dice cubes.

        # Disable the "Send Selected Dice Cubes" button
        print("Sending selected dice cubes:", selected_dice_cubes)
        self.send_button.config(state="disabled")

        self.controller.show_page("RunningGamePage")

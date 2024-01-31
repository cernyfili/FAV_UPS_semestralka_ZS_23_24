# import tkinter as tk
# from tkinter import messagebox
# import random
# import time
#
#
# class DiceGameGUI:
#     def __init__(self, master, player_names):
#         self.master = master
#         self.master.title("Dice 10000 Game")
#
#         self.players = player_names
#         self.current_player_index = 0
#
#         self.roll_button = tk.Button(master, text="Roll Dice", command=self.roll_dice)
#         self.roll_button.pack(pady=10)
#
#         self.player_label = tk.Label(master, text=f"Current Player: {self.players[self.current_player_index]}")
#         self.player_label.pack()
#
#         self.result_label = tk.Label(master, text="Roll result will be displayed here")
#         self.result_label.pack(pady=10)
#
#         self.canvas = tk.Canvas(master, width=400, height=150, bg="white")
#         self.canvas.pack()
#
#         self.cube_size = 40
#         self.dot_radius = 5
#
#         self.update_player_label()
#
#     def update_player_label(self):
#         self.player_label.config(text=f"Current Player: {self.players[self.current_player_index]}")
#
#     def roll_dice(self):
#         # Simulate rolling dice and get the result
#         dice_result = [random.randint(1, 6) for _ in range(6)]
#
#         # Display the result
#         self.result_label.config(text=f"Dice result: {dice_result}")
#
#         # Draw the rolled dice
#         self.draw_dice(dice_result)
#
#         # Check for game end or switch to the next player (placeholder logic)
#         if self.current_player_index < len(self.players) - 1:
#             self.current_player_index += 1
#         else:
#             self.current_player_index = 0
#             messagebox.showinfo("Game Over", "All players have completed their turns!")
#
#         self.update_player_label()
#
#     def draw_dice(self, dice_result):
#         # Clear previous drawings on the canvas
#         self.canvas.delete("all")
#
#         # Draw six red squares with dots
#         for i, number in enumerate(dice_result):
#             x, y = 50 + i * 60, 75
#             self.draw_square(x, y, number)
#
#         self.master.update()
#         time.sleep(1)
#
#     def draw_square(self, x, y, dots):
#         # Draw red square
#         self.canvas.create_rectangle(x - self.cube_size / 2, y - self.cube_size / 2,
#                                      x + self.cube_size / 2, y + self.cube_size / 2,
#                                      fill="red")
#
#         # Draw dots within the square
#         dot_positions = [(x - 10, y - 10), (x, y - 10), (x + 10, y - 10),
#                          (x - 10, y), (x, y), (x + 10, y)]
#
#         for dot_x, dot_y in dot_positions[:dots]:
#             self.canvas.create_oval(dot_x - self.dot_radius, dot_y - self.dot_radius,
#                                     dot_x + self.dot_radius, dot_y + self.dot_radius, fill="white")
#
#
# if __name__ == "__main__":
#     root = tk.Tk()
#
#     # Replace ["Player1", "Player2"] with the actual list of player names
#     players = ["Player1", "Player2"]
#
#     game_gui = DiceGameGUI(root, players)
#
#     root.mainloop()

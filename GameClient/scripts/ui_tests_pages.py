#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: ui_tests_pages.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
from frontend.ui_manager import MyApp
from shared.data_structures import GameList, Game, PlayerList, Player, GameData, PlayerGameData


def lobby_page():

    app = init_app()
    game_list = GameList( [Game("Game1", 2, 4), Game("Game2", 3, 5), Game("Game3", 1, 3)] )
    app.show_page_with_data("LobbyPage", game_list)
    app.mainloop()

def before_game_page():
    app = init_app()
    player_list = PlayerList( [Player("Player1")] )
    app.show_page_with_data("BeforeGamePage", player_list)
    app.mainloop()

def running_game_page():
    app = init_app()
    game_data = GameData([PlayerGameData("Karel", True, 5, True),PlayerGameData("Jan", False, 5, False)])
    app.show_page_with_data("RunningGamePage", game_data)
    app.mainloop()

def my_turn_page():
    app = init_app()
    app.show_page("MyTurnPage")
    app.mainloop()

def init_app():
    app = MyApp()
    app.geometry("800x400")
    return app


if __name__ == "__main__":
    my_turn_page()
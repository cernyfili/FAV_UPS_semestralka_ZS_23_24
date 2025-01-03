import os
import sys

from frontend.ui_manager import MyApp
from shared.data_structures import GameList, Game


def main():

    app = MyApp()
    app.geometry("800x400")
    game_list = GameList( [Game("Game1", 2, 4), Game("Game2", 3, 5), Game("Game3", 1, 3)] )
    app.show_page_with_data("LobbyPage", game_list)
    app.mainloop()


if __name__ == "__main__":
    main()
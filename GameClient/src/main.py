import logging
import os
import sys

from frontend.ui_manager import MyApp
from shared.constants import Game, GameList


def main():

    app = MyApp()
    app.geometry("800x400")
    app.show_page("StartPage")
    app.mainloop()


if __name__ == "__main__":
    # log to console
    logging.basicConfig(level=logging.DEBUG)
    main()
    # try:
    #     main()
    # except Exception as e:
    #     print(f"Error: {e}")
    #     sys.exit(1)
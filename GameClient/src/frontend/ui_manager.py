import tkinter as tk

from frontend.views.before_game_page import BeforeGamePage
from frontend.views.lobby_page import LobbyPage
from frontend.views.my_turn_page import MyTurnPage
from frontend.views.running_game_page import RunningGamePage
from frontend.views.start_page import StartPage


class MyApp(tk.Tk):
    def __init__(self):
        tk.Tk.__init__(self)
        self.title("Multi-Page Tkinter App")

        # Create container to hold pages
        self.container = tk.Frame(self)
        self.container.pack(side="top", fill="both", expand=True)
        self.container.grid_rowconfigure(0, weight=1)
        self.container.grid_columnconfigure(0, weight=1)
        self._current_frame = "StartPage"

        # Dictionary to hold pages
        self.pages = {}

        # Create pages
        for PageClass in (StartPage, LobbyPage, BeforeGamePage, RunningGamePage, MyTurnPage):
            page_name = PageClass.__name__
            frame = PageClass(parent=self.container, controller=self)
            self.pages[page_name] = frame
            frame.grid(row=0, column=0, sticky="nsew")

        # Show the initial page
        self.show_page("StartPage")

    def show_page(self, page_name, data=None):
        # Show the selected page
        self._current_frame = page_name
        frame = self.pages[page_name]
        if data:
            frame.update_data(data)

        frame.tkraise()
        # try:

        #     frame.tkraise()
        # except Exception as e:
        #     print(f"Error: {e}")
        #     # exit
        #     sys.exit(1)

    def get_current_frame(self):
        return self._current_frame



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

    def show_page(self, page_name):
        # Show the selected page
        frame = self.pages[page_name]
        frame.tkraise()

    def show_page_with_data(self, page_name, data):
        frame = self.pages[page_name]
        frame.update_data(data)
        frame.tkraise()


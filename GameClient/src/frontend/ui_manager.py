import tkinter as tk
from dataclasses import dataclass

from src.frontend.views.before_game_page import BeforeGamePage
from src.frontend.views.lobby_page import LobbyPage
from src.frontend.views.my_turn_roll_dice_page import MyTurnRollDicePage
from src.frontend.views.my_turn_select_cubes_page import MyTurnSelectCubesPage
from src.frontend.views.running_game_page import RunningGamePage
from src.frontend.views.start_page import StartPage

START_PAGE = StartPage.__name__

@dataclass
class PageData:
    page_name: str
    class_value: any

pages: dict[str, PageData] = {
    "StartPage": PageData(StartPage.__name__, StartPage),
    "LobbyPage": PageData(LobbyPage.__name__, LobbyPage),
    "BeforeGamePage": PageData(BeforeGamePage.__name__, BeforeGamePage),
    "RunningGamePage": PageData(RunningGamePage.__name__, RunningGamePage),
    "MyTurnRollDicePage": PageData(MyTurnRollDicePage.__name__, MyTurnRollDicePage),
    "MyTurnSelectCubesPage": PageData(MyTurnSelectCubesPage.__name__, MyTurnSelectCubesPage)
}



class MyApp(tk.Tk):
    def __init__(self):
        tk.Tk.__init__(self)
        self.title("Multi-Page Tkinter App")

        # Create container to hold pages
        self.container = tk.Frame(self)
        self.container.pack(side="top", fill="both", expand=True)
        self.container.grid_rowconfigure(0, weight=1)
        self.container.grid_columnconfigure(0, weight=1)
        self._current_frame = START_PAGE

        # Dictionary to hold pages
        self.pages = {}

        # Create pages
        # page_list is all CPageEnum.class_value
        page_list = [page.class_value for page in pages.values()]
        for PageClass in page_list:
            page_name = PageClass.__name__
            frame = PageClass(parent=self.container, controller=self)
            self.pages[page_name] = frame
            frame.grid(row=0, column=0, sticky="nsew")

        # Show the initial page
        self.show_page(START_PAGE)

    def show_page(self, page_name, data=None):
        if page_name not in self.pages:
            assert False, f"Page {page_name} not found"
        # Show the selected page
        self._current_frame = page_name
        frame = self.pages[page_name]

        frame.tkraise()

        if data:
            frame.update_data(data)

    def get_current_frame(self):
        return self._current_frame



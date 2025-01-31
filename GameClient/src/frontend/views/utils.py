#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: utils.py
Author: Filip Cerny
Created: 05.01.2025
Version: 1.0
Description: 
"""
import logging
import threading
from tkinter import messagebox
from types import SimpleNamespace

from src.backend.server_communication import ServerCommunication
from src.shared.constants import GAME_STATE_MACHINE, CCommandTypeEnum, Command, NetworkMessage, MessageFormatError, \
    MessageStateError

pages_names = {
    "StartPage": "StartPage",
    "LobbyPage": "LobbyPage",
    "BeforeGamePage": "BeforeGamePage",
    "RunningGamePage": "RunningGamePage",
    "MyTurnRollDicePage": "MyTurnRollDicePage",
    "MyTurnSelectCubesPage": "MyTurnSelectCubesPage",
}

PAGES_DIC = SimpleNamespace(**pages_names)

def process_is_not_connected() -> tuple[str, list | None]:
    def __inn_not_connected():
        next_page_name = PAGES_DIC.StartPage
        messagebox.showerror("Connection Failed", "Could not connect to the server")
        return next_page_name, None
    try:
        is_connected, received_message, message_list = ServerCommunication().communication_reconnect_message()
    except MessageFormatError  or MessageStateError  as e:
        ServerCommunication().close_connection()
        messagebox.showerror("Error", "Wrong message format")
        return PAGES_DIC.StartPage, None

    except Exception as e:
        logging.error(f"Error while trying to reconnect: {e}")
        if AssertionError:
            raise e
        return __inn_not_connected()

    if not is_connected:
        return __inn_not_connected()

    try:
        next_page_name = get_connect_next_page(received_message)
    except Exception as e:
        assert False, f"Unknown command: {received_message}"

    return next_page_name, message_list

def get_connect_next_page(message : NetworkMessage) -> str:
    next_page_name_dic: dict[int, str] = {
        CCommandTypeEnum.ResponseServerReconnectBeforeGame.value.id: PAGES_DIC.BeforeGamePage,
        CCommandTypeEnum.ResponseServerReconnectRunningGame.value.id: PAGES_DIC.RunningGamePage,
        CCommandTypeEnum.ResponseServerGameList.value.id: PAGES_DIC.LobbyPage
    }

    next_page_name = next_page_name_dic.get(message.command_id)
    if not next_page_name:
        raise Exception("Unknown command")
    return next_page_name

def stop_update_thread(self):
    logging.debug("Stopping the update thread")
    # region THREAD STOP
    self._stop_event.set()
    # Wait for the update thread to finish
    if self._update_thread.is_alive():
        logging.debug("Waiting for the update thread to finish")
        self._update_thread.join()
    # endregion


# region Loading animation
def stop_animation(self):
    if hasattr(self, 'loading_label'):
        self.loading_label.destroy()


def bind_to_enter_key(self, button):
    pass
    # self.bind('<Return>', button.invoke())


def animate(self, waiting_animation, label_str):
    if not waiting_animation.winfo_exists():
        return

    text = waiting_animation.cget("text")
    if text.count(".") < 3:
        text += "."
    else:
        text = label_str

    waiting_animation.config(text=text)
    self.after(500, lambda: animate(self, waiting_animation, label_str))


def show_loading_animation(self, tk):
    # destroy the current content
    for widget in self.winfo_children():
        widget.destroy()

    self.loading_label = tk.Label(self, text="Loading...")
    self.loading_label.pack(pady=10, padx=10)
    animate(self, self.loading_label, "Loading")

def show_reconnecting_animation(self, tk):
    # destroy the current content
    for widget in self.winfo_children():
        widget.destroy()

    self.loading_label = tk.Label(self, text="Attempting to reconnect...")
    self.loading_label.pack(pady=10, padx=10)
    animate(self, self.loading_label, "Attempting to reconnect")


# endregion

# def _standard_start_listening_for_updates(self, process_command, listen_for_updates: callable, continue_commands):
#     pass


def list_start_listening_for_updates(self, process_command_dic: dict[int, callable], update_command: Command,
                                     continue_commands: list[Command]):
    def listen_for_updates(state_name: str, update_function: callable, process_command_dic: dict[int, callable],
                           update_command: Command, continue_commands_list: list[Command], stop_event: threading.Event):
        logging.debug("Listening for updates")
        current_state = GAME_STATE_MACHINE.get_current_state()
        while current_state == state_name and not stop_event.is_set():
            logging.debug("Listening for Data updates")
            try:
                if stop_event.is_set():
                    return
                try:
                    is_connected, message_info_list = update_function()
                except MessageFormatError  or MessageStateError  as e:
                    self._show_wrong_message_format()

                if stop_event.is_set():
                    return

                if not is_connected:
                    self._show_process_is_not_connected()
                    return

                if not message_info_list:
                    # timeout trying read update message
                    continue

                for i, message_info in enumerate(message_info_list):
                    if stop_event.is_set():
                        return

                    command, message_data = message_info

                    is_handled = False
                    for continue_command in continue_commands_list:
                        if command.id != continue_command.id:
                            continue
                        is_handled = True
                        break
                    if is_handled:
                        continue

                    if command == update_command:
                        self.update_data(message_data)
                        continue

                    for command_id, handler in process_command_dic.items():
                        if command_id != command.id:
                            continue
                        # check if handler requires arguments
                        if handler.__code__.co_argcount > 1:
                            handler(message_data)
                        else:
                            handler()
                        return

                    raise Exception("Unknown command or in buffer message for different page")
            except Exception as e:
                raise e
                # todo change
                messagebox.showerror("Error", str(e))
                break

    state_name = self._get_state_name()
    update_function = self._get_update_function()

    stop_event = threading.Event()
    self._stop_event = stop_event


    logging.debug("Starting to listen for updates")
    # Wait for the update thread to finish
    self._set_update_thread(
        threading.Thread(target=listen_for_updates,
                         args=(state_name, update_function, process_command_dic, update_command, continue_commands,
                               stop_event), daemon=True))
    self._update_thread.start()

    # if command == CCommandTypeEnum.ServerPingPlayer.value:
    #     continue
    #
    # if command == CCommandTypeEnum.ServerUpdateGameData.value:
    #     self.update_data(message_data)
    #     continue


def start_listening_for_updates_update_gamedata(self):
    process_command = {
        CCommandTypeEnum.ResponseServerError.value.id: self.process_error,
        CCommandTypeEnum.ServerUpdateNotEnoughPlayers.value.id: self._process_update_not_enough_players
    }
    continue_commands = [CCommandTypeEnum.ServerPingPlayer.value]
    update_command = CCommandTypeEnum.ServerUpdateGameData.value
    list_start_listening_for_updates(self, process_command, update_command, continue_commands)


def destroy_elements(self):
    if self.winfo_children():
        for widget in self.winfo_children():
            if widget.winfo_exists():
                try:
                    widget.destroy()
                except Exception as e:
                    logging.error(f"Error while destroying widget: {e}")

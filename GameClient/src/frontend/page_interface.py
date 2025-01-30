#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: page_interface.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import threading
import tkinter as tk
from abc import ABC, abstractmethod

from src.backend.server_communication import ServerCommunication
from src.frontend.views.utils import stop_update_thread, PAGES_DIC, process_is_not_connected, show_loading_animation, \
    stop_animation, show_reconnecting_animation


#
# class PageInterface(ABC):
#     @abstractmethod
#     def _load_page_content(self):
#         pass
#
#     @abstractmethod
#     def update_data(self, data):
#         pass

class UpdateInterface(ABC):
    def __init__(self):
        self.controller = None
        self._stop_event = None
        self._update_thread = None
        self._list = None
        self._lock = None
        self._is_connected = None

    @abstractmethod
    def tkraise(self, aboveThis=None):
        pass

    def _show_logout_button(self, tk):
        # Logout button in right upper corner of window
        logout_button = tk.Button(self, text="Logout", command=lambda: self._button_action_logout())
        logout_button.place(relx=1.0, rely=0.0, anchor="ne")

    def _button_action_logout(self):
        send_function = ServerCommunication().send_client_logout
        next_page_name = PAGES_DIC.StartPage
        page_data = None

        stop_update_thread(self)

        try:
            is_connected = send_function()
            if not is_connected:
                next_page_name, page_data = process_is_not_connected()
        except Exception as e:
            next_page_name, page_data = process_is_not_connected()

        self.controller.show_page(next_page_name, page_data)

    # def _start_listening_for_updates(self):
    #     state_name = self._get_state_name()
    #     list_update_function = self._get_update_function()
    #
    #     logging.debug(f"Starting to listen for updates in state: {state_name}")
    #
    #     def listen_for_updates(state_name, list_update_function):
    #         current_state = GAME_STATE_MACHINE.get_current_state()
    #         while current_state == state_name and not self._stop_event.is_set():
    #             logging.debug("Listening for list updates")
    #             try:
    #                 is_connected, received_list = list_update_function()
    #                 if self._stop_event.is_set():
    #                     break
    #                 if not is_connected:
    #                     process_is_not_connected(self)
    #                     break
    #                 if not received_list:
    #                     continue
    #                 self.update_data(received_list)
    #                 continue
    #             except Exception as e:
    #                 raise e
    #                 #todo change
    #                 messagebox.showerror("Error", str(e))
    #                 break
    #
    #     # Wait for the update thread to finish
    #     self._set_update_thread(
    #         threading.Thread(target=listen_for_updates, args=(state_name, list_update_function), daemon=True))
    #     self._update_thread.start()

    def button_action_standard(self, tk, send_function, next_page_name, param_list):
        def run_send_function(send_function, next_page_name, param_list):

            stop_update_thread(self)
            page_data = None
            try:
                is_connected = send_function(param_list)
                if not is_connected:
                    next_page_name, page_data = process_is_not_connected()
            except Exception as e:
                next_page_name, page_data = process_is_not_connected()

            stop_animation(self)
            self.controller.show_page(next_page_name, page_data)

        show_loading_animation(self, tk)

        threading.Thread(target=run_send_function, args=(send_function, next_page_name, param_list)).start()
        return True

    @abstractmethod
    def _get_state_name(self):
        pass

    @abstractmethod
    def _get_update_function(self):
        pass

    def update_data(self, data):
        with self._lock:  # Acquire the lock
            self._list = data
            self._load_page_content()

    def _show_process_is_not_connected(self):
        with self._lock:
            def run_send_function():
                next_page_name, page_data = process_is_not_connected()
                stop_animation(self)
                self.controller.show_page(next_page_name, page_data)

            show_reconnecting_animation(self, tk)
            threading.Thread(target=run_send_function, args=()).start()


    @abstractmethod
    def _load_page_content(self):
        pass

    @abstractmethod
    def _set_update_thread(self, param):
        pass

    # set is connected
    def set_is_connected(self, is_connected):
        with self._lock:
            self._is_connected = is_connected

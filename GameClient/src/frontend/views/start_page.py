#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: start_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import logging
import re
import threading
import time
import tkinter as tk

from src.backend.server_communication import ServerCommunication
from src.frontend.views.utils import process_is_not_connected
from src.frontend.views.utils import show_loading_animation, stop_animation, get_connect_next_page, \
    destroy_elements
from src.shared.constants import CMessageConfig


class StartPage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        self._load_page_content()

    def tkraise(self, aboveThis=None):
        page_name = self.winfo_name()
        logging.debug(f"Raising Page: {page_name}")
        # Call the original tkraise method
        super().tkraise(aboveThis)

        # Custom behavior after raising the frame
        self._load_page_content()

    def _load_page_content(self):
        destroy_elements(self)


        label = tk.Label(self, text="Connect to a Server")
        label.pack(pady=10, padx=10, fill='both', expand=True)

        # Form with IP and Port and NickName
        ip_label = tk.Label(self, text="IP Address")
        ip_label.pack(pady=10, padx=10, fill='both', expand=True)
        ip_entry_var = tk.StringVar(value="192.168.0.31")
        ip_entry = tk.Entry(self, textvariable=ip_entry_var, validate="focusout", validatecommand=(self.register(self._validate_ip), '%P'))
        ip_entry.pack(pady=10, padx=10, fill='both', expand=True)

        port_label = tk.Label(self, text="Port")
        port_label.pack(pady=10, padx=10, fill='both', expand=True)
        port_entry_var = tk.StringVar(value="10000")
        port_entry = tk.Entry(self, textvariable=port_entry_var, validate="focusout", validatecommand=(self.register(self._validate_port), '%P'))
        port_entry.pack(pady=10, padx=10, fill='both', expand=True)

        nickname_label = tk.Label(self, text="Nickname")
        nickname_label.pack(pady=10, padx=10, fill='both', expand=True)

        randomNum = str(time.time())
        player_str = "Player" + randomNum[-5:]
        # player_str = "Player1"

        nickname_entry_var = tk.StringVar(value=player_str)
        nickname_entry = tk.Entry(self, textvariable=nickname_entry_var, validate="focusout", validatecommand=(self.register(self._validate_nickname), '%P'))
        nickname_entry.pack(pady=10, padx=10, fill='both', expand=True)

        connect_button = tk.Button(self, text="Connect",
                                   command=lambda: self._button_action_connect(
                                       ip_entry.get(), int(port_entry.get()), nickname_entry.get())
                                   )
        connect_button.pack(pady=10, padx=10, fill='both', expand=True)

        self.bind('<Return>',
                  lambda event: self._button_action_connect(ip_entry.get(), int(port_entry.get()), nickname_entry.get()))

    @staticmethod
    def _validate_ip(ip):
        # Check if ip is a valid IP address
        return bool(re.match(r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$', ip))

    @staticmethod
    def _validate_port(port):
        # Check if port is a number and within the valid range
        return port.isdigit() and 1 <= int(port) <= 65535

    @staticmethod
    def _validate_nickname(nickname):
        return CMessageConfig.is_valid_name(nickname)

    def _button_action_connect(self, ip : str, port : int, nickname : str):

        def run_send_function(ip, port, nickname):
            try:
                is_connected, response_message, update_list = ServerCommunication().send_login_message(ip, port,
                                                                                                       nickname)
            except Exception as e:
                next_page_name, page_data = process_is_not_connected()
                stop_animation(self)
                self.controller.show_page(next_page_name, page_data)
                return

            if not is_connected:
                next_page_name, page_data = process_is_not_connected()
                stop_animation(self)
                self.controller.show_page(next_page_name, page_data)
                return

            try:
                next_page_name = get_connect_next_page(response_message)
            except Exception as e:
                assert False, f"Unknown command: {response_message}"

            stop_animation(self)
            self.controller.show_page(next_page_name, update_list)

        show_loading_animation(self, tk)

        threading.Thread(target=run_send_function, args=(ip, port, nickname)).start()
        return True

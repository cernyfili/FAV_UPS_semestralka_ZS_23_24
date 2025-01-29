#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: start_page.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import re
import threading
import time
import tkinter as tk

from src.backend.server_communication import ServerCommunication
from src.frontend.views.utils import PAGES_DIC, show_loading_animation, stop_loading_animation
from src.frontend.views.utils import process_is_not_connected
from src.shared.constants import CMessageConfig, CCommandTypeEnum


class StartPage(tk.Frame):
    def __init__(self, parent, controller):
        tk.Frame.__init__(self, parent)
        self.controller = controller

        label = tk.Label(self, text="Connect to a Server")
        label.pack(pady=10, padx=10, fill='both', expand=True)

        # Form with IP and Port and NickName
        ip_label = tk.Label(self, text="IP Address")
        ip_label.pack(pady=10, padx=10, fill='both', expand=True)
        ip_entry_var = tk.StringVar(value="127.0.0.1")
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
            next_page_name_dic: dict[int, str] = {
                CCommandTypeEnum.ResponseServerReconnectBeforeGame.value.id: PAGES_DIC.BeforeGamePage,
                CCommandTypeEnum.ResponseServerReconnectRunningGame.value.id: PAGES_DIC.RunningGamePage,
                CCommandTypeEnum.ResponseServerGameList.value.id: PAGES_DIC.LobbyPage
            }

            try:
                is_connected, response_command, update_list = ServerCommunication().send_connect_message(ip, port,
                                                                                                         nickname)
                if not is_connected:
                    process_is_not_connected(self)
            except Exception as e:
                process_is_not_connected(self)
                # todo remove
                raise
            finally:
                stop_loading_animation(self)

            next_page_name = None
            for key, value in next_page_name_dic.items():
                if response_command.command_id != key:
                    continue
                next_page_name = value
                break
            if not next_page_name:
                raise Exception("Unknown command")

            self.controller.show_page(next_page_name, update_list)

        show_loading_animation(self, tk)

        threading.Thread(target=run_send_function, args=(ip, port, nickname)).start()
        return True

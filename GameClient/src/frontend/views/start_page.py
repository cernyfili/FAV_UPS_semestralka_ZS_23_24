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
import tkinter as tk
from tkinter import messagebox

from backend.server_communication import ServerCommunication
from frontend.views.utils import process_is_not_connected


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
        nickname_entry_var = tk.StringVar(value="Player1")
        nickname_entry = tk.Entry(self, textvariable=nickname_entry_var, validate="focusout", validatecommand=(self.register(self._validate_nickname), '%P'))
        nickname_entry.pack(pady=10, padx=10, fill='both', expand=True)

        connect_button = tk.Button(self, text="Connect",
                                   command=lambda: self._button_action_connect(
                                       ip_entry.get(), int(port_entry.get()), nickname_entry.get())
                                   )
        connect_button.pack(pady=10, padx=10, fill='both', expand=True)

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
        # Check if nickname is not empty
        return bool(nickname)

    def _button_action_connect(self, ip : str, port : int, nickname : str):

        try:
            is_connected, game_list = ServerCommunication().send_client_login(ip, port, nickname)
            if not is_connected:
                process_is_not_connected(self)
        except Exception as e:
            #messagebox.showerror("Connection Failed", str(e))
            #todo

            raise e
            return

        self.controller.show_page("LobbyPage", game_list)



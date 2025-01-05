#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: utils.py
Author: Filip Cerny
Created: 05.01.2025
Version: 1.0
Description: 
"""
from tkinter import messagebox


def process_is_not_connected(instance_object):
    # todo messagebox.showerror("Connection Failed", "Could not connect to the server")
    raise ConnectionError("Could not connect to the server")
    instance_object.controller.show_page("StartPage")

def stop_update_thread(self):
    # region THREAD STOP
    self._stop_event.set()
    # Wait for the update thread to finish
    if self._update_thread.is_alive():
        self._update_thread.join()
    else:
        raise ValueError("Update thread not found")
    # endregion
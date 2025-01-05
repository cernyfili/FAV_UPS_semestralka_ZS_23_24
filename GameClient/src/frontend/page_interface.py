#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: page_interface.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
import logging
import threading
from abc import ABC, abstractmethod

from shared.constants import GAME_STATE_MACHINE


class PageInterface(ABC):
    @abstractmethod
    def _load_page_content(self):
        pass

    @abstractmethod
    def update_data(self, data):
        pass

class UpdateInterface(ABC):
    def __init__(self):
        self._stop_event = None
        self._update_thread = None
        self._list = None
        self._lock = None

    @abstractmethod
    def tkraise(self, aboveThis=None):
        pass

    def _start_listening_for_updates(self):
        state_name = self._get_state_name()
        update_function = self._get_update_function()

        logging.debug("Starting to listen for updates")
        def listen_for_updates():
            current_state = GAME_STATE_MACHINE.get_current_state()
            while current_state == state_name and not self._stop_event.is_set():
                logging.debug("Listening for game list updates")
                try:
                    update_list = update_function()
                    if self._stop_event.is_set():
                        break
                    if not update_list:
                        break
                    self.update_data(update_list)
                except Exception as e:
                    raise e
                    break
        # Wait for the update thread to finish
        self._set_update_thread(threading.Thread(target=listen_for_updates, daemon=True))
        self._update_thread.start()

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

    @abstractmethod
    def _load_page_content(self):
        pass

    @abstractmethod
    def _set_update_thread(self, param):
        pass
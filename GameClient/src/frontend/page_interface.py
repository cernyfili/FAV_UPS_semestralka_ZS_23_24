#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: page_interface.py
Author: Filip Cerny
Created: 02.01.2025
Version: 1.0
Description: 
"""
from abc import ABC, abstractmethod


class PageInterface(ABC):
    @abstractmethod
    def _load_page_content(self):
        pass

    @abstractmethod
    def update_data(self, data):
        pass
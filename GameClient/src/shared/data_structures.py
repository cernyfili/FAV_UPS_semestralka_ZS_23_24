#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: data_structures.py
Author: Filip Cerny
Created: 01.01.2025
Version: 1.0
Description: 
"""
from dataclasses import dataclass
from typing import List, TypeAlias

import timestamp

from shared.constants import CMessageConfig, CCommandTypeEnum

@dataclass
class MessageParamListInfo:
    param_names: list[str]
    convert_function: callable = None


@dataclass
class Game:
    name: str
    connected_count_players: int
    max_players: int

GameList: TypeAlias = list[Game]

@dataclass
class Player:
    name: str

PlayerList: TypeAlias = list[Player]

@dataclass
class PlayerGameData:
    player_name: str
    is_connected: bool
    score: int
    is_turn: bool

GameData: TypeAlias = list[PlayerGameData]


@dataclass
class ScoreCube:
    Value: int
    ScoreValue: int


@dataclass
class Param:
    name: str
    value: any


@dataclass
class Command:
    id: int
    trigger: callable
    param_names: List[str]
    param_list_info: any

    def is_valid_param_names(self, params: List[Param]) -> bool:
        param_names = [param.name for param in params]
        return self.param_names == param_names


@dataclass(frozen=True)
class Brackets:
    opening: str
    closing: str


@dataclass
class NetworkMessage:

    def __init__(self, signature: str, command_id: int, timestamp: timestamp, player_nickname: str, parameters: List[Param]):
        def process_signature(signature: str) -> str:
            if signature != CMessageConfig.SIGNATURE:
                raise ValueError("Invalid signature")
            return signature

        def process_command_id(command_id: int) -> int:
            if not CCommandTypeEnum.is_command_id_in_enum(command_id):
                raise ValueError("Invalid command ID")
            return command_id

        def process_player_nickname(player_nickname: str) -> str:
            if not CMessageConfig.is_valid_name(player_nickname):
                raise ValueError("Invalid player nickname")
            return player_nickname

        self._signature : str = process_signature(signature)
        self._command_id : int = process_command_id(command_id)
        self._timestamp : timestamp = timestamp
        self._player_nickname : str = process_player_nickname(player_nickname)
        self._parameters : List[Param] = self._process_parameters(parameters)

    def _process_parameters(self, parameters: List[Param]) -> List[Param]:
        if CCommandTypeEnum.is_command_id_in_enum(self._command_id):
            command = CCommandTypeEnum.get_command_by_id(self._command_id)
            if not command.is_valid_param_names(parameters):
                raise ValueError("Invalid parameters")
        return parameters

    @property
    def signature(self) -> str:
        return self._signature

    @property
    def command_id(self) -> int:
        return self._command_id

    @property
    def timestamp(self) -> timestamp:
        return self._timestamp

    @property
    def player_nickname(self) -> str:
        return self._player_nickname

    @property
    def parameters(self) -> List[Param]:
        return self._parameters


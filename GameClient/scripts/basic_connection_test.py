#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File name: basic_connection_test.py
Author: Filip Cerny
Created: 31.12.2024
Version: 1.0
Description: 
"""

import socket

# Define server address and port
server_address = ('localhost', 10000)

# Create a TCP/IP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

try:
    # Connect to the server
    sock.connect(server_address)
    print(f"Connected to server at {server_address}")

    # Send a message to the server
    messageStr = ""
    messageStr += "KIVUPS"
    messageStr += "01"
    messageStr += "2024-12-31T15:30:00Z"
    messageStr += "{nickname}"
    messageStr += "{}"
    messageStr += "\n"
    sock.sendall(messageStr.encode())  # Send the message

    while True:
        # Optionally, receive a response from the server
        response = sock.recv(1024)  # Receive up to 1024 bytes of data
        if not response:
            continue

        response_str = response.decode()
        print(f"Received response from server: {response_str}")
        # if response_str == "KIVUPS332024-12-31T15:30:00Z{nickname}{\"gameList\":\"[]\",}":
        send_message = "KIVUPS022024-12-31T15:30:00Z{nickname}{\"gameName\":\"game_name\",maxPlayers:\"2\"}"
        sock.sendall(send_message.encode())

finally:
    # Close the connection
    sock.close()
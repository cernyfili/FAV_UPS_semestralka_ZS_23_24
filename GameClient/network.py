import socket

def send_and_receive(server_ip, server_port, message):
    # Create a socket object
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    # Connect to the server
    s.connect((server_ip, server_port))

    # Send data to the server
    s.sendall(message.encode())

    # Receive data from the server
    data = s.recv(1024)

    # Close the connection
    s.close()

    return data.decode()
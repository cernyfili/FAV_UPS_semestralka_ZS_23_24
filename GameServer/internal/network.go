package internal

import (
	"fmt"
	"net"
	"os"
)

// StartServer starts a TCP server that listens on the specified port.
func StartServer(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("Server is listening on port " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		conn.Close()
		return
	}
	message := string(buffer)
	fmt.Println("Received: " + message)

	// Respond to the client
	_, err = conn.Write([]byte("Message received"))
	if err != nil {
		fmt.Println("Error writing:", err)
	}

	conn.Close()
}

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

/*package main

import (
	"fmt"
	"net"
	"os"
)

// CustomConnection represents a custom type that includes net.Conn and additional data.
type CustomConnection struct {
	net.Conn
	// Add any additional data fields you need
	UserID int
}

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

		// Create a CustomConnection with additional data
		customConn := CustomConnection{
			Conn:   conn,
			UserID: 123, // Add any data you want to associate with the connection
		}

		go handleConnection(customConn)
	}
}

func handleConnection(customConn CustomConnection) {
	buffer := make([]byte, 1024)
	_, err := customConn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		customConn.Close()
		return
	}
	message := string(buffer)
	fmt.Printf("Received from UserID %d: %s\n", customConn.UserID, message)

	// Process the received message and generate a response
	response := processMessage(message)

	// Respond to the client with the generated response
	_, err = customConn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing:", err)
	}

	customConn.Close()
}

func processMessage(message string) string {
	// Your logic for processing the received message and generating a response
	// For example, you can switch on the message content and determine the response.
	switch message {
	case "Hello":
		return "Hi there!"
	case "How are you?":
		return "I'm doing well, thank you!"
	default:
		return "Unknown message"
	}
}

func main() {
	StartServer("8080")
}
*/

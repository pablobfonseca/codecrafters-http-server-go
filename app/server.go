package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Initializating server...")

	connection := NewConnection("tcp", "localhost", 4221)
	listener, err := connection.connect()

	if err != nil {
		fmt.Printf("Error establishing a connection: %s\n", err.Error())
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go connection.handle(conn)
	}
}

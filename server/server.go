package main

import (
	"fmt"
	"net"
	"os"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type Server struct {
	listener net.Listener
	clients  map[string]net.Conn
}

type Client struct {
	conn       net.Conn
	p2p_client chan net.Conn
}

func main() {
	server := Server{}
	fmt.Println("Server Running...")
	listener, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	server.listener = listener
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.listener.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	for {
		connection, err := server.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("client connected")
		client := Client{connection, make(chan net.Conn)}
		go server.processClient(client)
	}
}
func (Server) processClient(client Client) {
	buffer := make([]byte, 1024)
	mLen, err := client.conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))

	_, err = client.conn.Write([]byte("test"))

	client.conn.Close()
}

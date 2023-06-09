package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = ""
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type Server struct {
	listener net.Listener
	clients  map[string]Client
}

type Client struct {
	conn       net.Conn
	p2p_client chan net.Conn
	open_chan  chan bool
}

func (server Server) processClient(client Client) {
	defer func() {
		client.conn.Close()
		delete(server.clients, strings.Split(client.conn.RemoteAddr().String(), ":")[0])
	}()

	buffer := make([]byte, 1024)
	mLen, err := client.conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))
	address := string(buffer[:mLen])
	var msg string
	if return_addr, ok := server.clients[address]; ok {
		fmt.Println("Punching through")
		msg = return_addr.conn.RemoteAddr().String() + " S"
		return_addr.p2p_client <- client.conn
		_, err = client.conn.Write([]byte(msg))

		mLen, err = client.conn.Read(buffer)
		if string(buffer[:mLen]) == "Success" {
			return_addr.open_chan <- true
		} else {
			return_addr.open_chan <- false
		}
	} else {
		fmt.Println("waiting for connection to ask to be punched through")
		return_conn := <-client.p2p_client
		msg = return_conn.RemoteAddr().String() + " C"
		_, err = client.conn.Write([]byte(msg))
		_, err = client.conn.Read(buffer)
		success := <-client.open_chan

		if success {
			client.conn.Write([]byte("Server Opened"))
		} else {
			client.conn.Write([]byte("F"))
		}
	}

}

func main() {
	server := Server{nil, make(map[string]Client)}
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
		fmt.Println("client connected, " + connection.RemoteAddr().String() + "  " + connection.LocalAddr().Network())
		client := Client{connection, make(chan net.Conn), make(chan bool)}
		server.clients[strings.Split(connection.RemoteAddr().String(), ":")[0]] = client
		go server.processClient(client)
	}
}

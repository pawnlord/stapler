package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = "192.168.1.19"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type ConnType string

const (
	ServerConn ConnType = "S"
	ClientConn ConnType = "C"
)

func strToConnType(str string) ConnType {
	switch str {
	case "S":
		return ServerConn
	case "C":
		return ClientConn
	}
}

func main() {
	var p2p_addr string
	var conn_type ConnType

	//establish connection
	connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}
	///send some data
	_, err = connection.Write([]byte("192.168.1.19"))
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))
	var conn_string string
	p2p_addr, conn_string = strings.Split(string(buffer[:mLen]), " ")[0], strings.Split(string(buffer[:mLen]), " ")[1]

	conn_type = strToConnType(conn_string)
	if conn_type == ServerConn {
		serverMain(p2p_addr, connection)
	} else {
		clientMain(p2p_addr, connection)
	}

}

func serverMain(p2p_addr string, original_server net.Conn) {
	var server net.Listener
	var err error
	{
		defer original_server.Close()
		server, err = net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
		if err != nil {
			original_server.Write([]byte("Fail"))
			fmt.Errorf(err.Error())
			return
		}
		original_server.Write([]byte("Success"))
	}
	defer server.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		if strings.Split(connection.RemoteAddr().String(), ":")[0] != strings.Split(p2p_addr, ":")[0] {
			fmt.Errorf("SECURITY WARNING: unexpected client connected, closing connection")
			connection.Close()
		}
		connection.Write([]byte("Hello!!"))
		connection.Close()
	}

}

func clientMain(p2p_addr string, original_server net.Conn) {
	var connection net.Conn

	buffer := make([]byte, 1024)
	{
		defer original_server.Close()
		original_server.Write([]byte("Success"))

		mLen, err := original_server.Read(buffer)
		if string(buffer[:mLen]) == "F" {
			fmt.Errorf("Server failed on remote client")
		}

		connection, err = net.Dial(SERVER_TYPE, p2p_addr)
		if err != nil {
			original_server.Write([]byte("Fail"))
		}

	}
	defer connection.Close()
	connection.Read(buffer)

	fmt.Println(string(buffer))
}

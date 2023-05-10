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
	ServerConn  ConnType = "S"
	ClientConn  ConnType = "C"
	InvalidConn ConnType = ""
)

func strToConnType(str string) ConnType {
	switch str {
	case "S":
		return ServerConn
	case "C":
		return ClientConn
	default:
		return InvalidConn
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
	conn_info := strings.Split(string(buffer[:mLen]), " ")

	p2p_addr = conn_info[0]
	conn_type = strToConnType(conn_info[1])

	if conn_type == ServerConn {
		serverMain(p2p_addr, connection)
	} else {
		clientMain(p2p_addr, connection)
	}
}

func serverMain(p2p_addr string, original_server net.Conn) {
	var server net.Listener
	defer original_server.Close()

	fmt.Println("Creating server from " + original_server.LocalAddr().String())

	server, err := net.Listen(SERVER_TYPE, original_server.LocalAddr().String())
	if err != nil {
		original_server.Write([]byte("Fail"))
		fmt.Println(err.Error())
		return
	}
	original_server.Write([]byte("Success"))

	defer server.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		if strings.Split(connection.RemoteAddr().String(), ":")[0] != strings.Split(p2p_addr, ":")[0] {
			fmt.Println("SECURITY WARNING: unexpected client connected, closing connection")
			connection.Close()
		}
		connection.Write([]byte("Hello!!"))
		connection.Close()
	}

}

func clientMain(p2p_addr string, original_server net.Conn) {
	var connection net.Conn
	fmt.Println("Creating client to " + p2p_addr)
	buffer := make([]byte, 1024)
	{
		defer original_server.Close()
		original_server.Write([]byte("Success"))

		mLen, err := original_server.Read(buffer)
		if string(buffer[:mLen]) == "F" {
			fmt.Println("Server failed on remote client")
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

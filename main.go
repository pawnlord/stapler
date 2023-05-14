package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pawnlord/stapler/nat"
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
	var dialer *net.Dialer

	dialer = nat.NewNATDialer()

	//establish connection
	connection, err := dialer.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
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
	var err error
	original_closed := false
	defer func() {
		if original_closed {
			original_server.Close()
		}
	}()
	fmt.Println("Starting server from " + original_server.LocalAddr().String())

	server, err = net.Listen(SERVER_TYPE, original_server.LocalAddr().String())
	if err != nil {
		original_server.Write([]byte("Fail"))
		fmt.Println(err.Error())
		return
	}
	original_server.Write([]byte("Success"))
	original_server.Close()

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
		} else {
			connection.Write([]byte("Hello!!---\x00"))
			connection.Close()
			return
		}
	}

}

func clientMain(p2p_addr string, original_server net.Conn) {
	var connection net.Conn
	fmt.Println("Creating client to " + p2p_addr)
	buffer := make([]byte, 1024)
	{
		defer original_server.Close()
		fmt.Println("Writing success to original server")
		original_server.Write([]byte("Success"))

		fmt.Println("Reading for failure from original server")
		mLen, err := original_server.Read(buffer)
		if string(buffer[:mLen]) == "F" {
			fmt.Println("Server failed on remote client")
			return
		}
		fmt.Println("Success, connecting to p2p server")

		connection, err = net.Dial(SERVER_TYPE, p2p_addr)
		if err != nil {
			fmt.Println("Fail, aborting")
			original_server.Write([]byte("Fail"))
			return
		}

	}
	defer connection.Close()
	fmt.Println("Success, reading from p2p server")
	connection.Read(buffer)

	fmt.Println(string(buffer))
}

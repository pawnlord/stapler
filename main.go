package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/pawnlord/stapler/nat"
)

const (
	SERVER_HOST = "192.168.1.19"
	SERVER_PORT = "9988"
)

func main() {
	var p2p_addr string
	var conn_type nat.ConnType
	conf := nat.ClientConfig{HostServer: SERVER_HOST, Port: SERVER_PORT}
	connection, err := nat.NewNATClient(conf)
	if err != nil {
		panic(err)
	}
	///send some data
	_, err = connection.MainConn.Write([]byte("192.168.1.19"))
	buffer := make([]byte, 1024)
	mLen, err := connection.MainConn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))
	conn_info := strings.Split(string(buffer[:mLen]), " ")

	connection.P2pAddr = conn_info[0]
	conn_type = nat.StrToConnType(conn_info[1])

	if conn_type == nat.LeadConn {
		serverMain(p2p_addr, connection)
	} else {
		clientMain(connection)
	}
}

func serverMain(p2p_addr string, client *nat.NATClient) {
	var server net.Listener
	var err error
	original_server := client.MainConn
	fmt.Println("Starting server from " + original_server.LocalAddr().String())

	server, err = net.Listen("tcp", original_server.LocalAddr().String())
	if err != nil {
		original_server.Write([]byte("Fail"))
		fmt.Println(err.Error())
		return
	}
	original_server.Write([]byte("Success"))

	defer server.Close()

	client.Accept()

	client.P2pConn.Write([]byte("Hello!!"))

}

func clientMain(client *nat.NATClient) {
	original_server := client.MainConn
	fmt.Println("Creating client to " + client.P2pAddr)
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

		client.P2pConn, err = net.Dial("tcp", client.P2pAddr)
		if err != nil {
			fmt.Println("Fail, aborting")
			original_server.Write([]byte("Fail"))
			return
		}

	}
	defer client.P2pConn.Close()
	fmt.Println("Success, reading from p2p server")
	client.P2pConn.Read(buffer)

	fmt.Println(string(buffer))
}

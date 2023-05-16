package nat

import (
	"fmt"
	"net"
	"strings"
)

type ConnType string

const (
	LeadConn     ConnType = "S"
	FollowerConn ConnType = "C"
	InvalidConn  ConnType = ""
)

type ClientConfig struct {
	HostServer string
	Port       string
}

type NATClient struct {
	dialer *net.Dialer
	// Our role in connecting. Either Server or Client
	Role ConnType
	// Connection with "match-making" server
	MainConn net.Conn
	// Connection with peer
	P2pConn net.Conn
	// Temporary server for the "lead" client, only used to establish communication
	Server net.Listener
	// Negotiated address we expect communication from
	P2pAddr string
}

func NewNATClient(conf ClientConfig) (*NATClient, error) {
	dialer := NewNATDialer()
	conn, err := dialer.Dial("tcp", conf.HostServer+":"+conf.Port)
	if err != nil {
		return nil, err
	}

	return &NATClient{dialer, InvalidConn, conn, nil, nil, ""}, nil
}

func (client *NATClient) Accept() error {
	if client.Role != LeadConn {
		return fmt.Errorf("Client is not acting as server, continuing")
	}
	if client.Server == nil {
		return fmt.Errorf("Client has not initialized listener")
	}
	if client.P2pAddr == "" {
		return fmt.Errorf("No known peer to accept")
	}
	for {
		connection, err := client.Server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		if strings.Split(connection.RemoteAddr().String(), ":")[0] != strings.Split(client.P2pAddr, ":")[0] {
			fmt.Println("SECURITY WARNING: unexpected client connected, closing connection")
			connection.Close()
		} else {
			client.P2pConn = connection
			return nil
		}
	}

}

type client struct {
	conn       net.Conn
	p2p_client chan net.Conn
	open_chan  chan bool
}

type ServerConfig struct {
	hostServer string
	port       string
}

type NATServer struct {
	net.Listener
	clients map[string]client
}

func newNATServer(conf ServerConfig) (*NATServer, error) {
	listener, err := net.Listen("tcp", conf.hostServer+":"+conf.port)
	if err != nil {
		return nil, err
	}
	return &NATServer{listener, make(map[string]client)}, nil
}

func NewClientEntry(conn net.Conn) client {
	return client{conn, make(chan net.Conn), make(chan bool)}
}

func Test() {
	fmt.Println("Hello!!!!")
}

func StrToConnType(str string) ConnType {
	switch str {
	case "S":
		return LeadConn
	case "C":
		return FollowerConn
	default:
		return InvalidConn
	}
}

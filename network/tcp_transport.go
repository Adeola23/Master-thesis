package network

import (
	_ "bytes"
	"encoding/gob"
	"fmt"
	_ "io"
	_ "log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type NetAddr string

func (n NetAddr) String() string  { return string(n) }
func (n NetAddr) Network() string { return "tcp" }

type Peer struct {
	conn         net.Conn
	listenAddr   string
	status    bool
	LastPingTime time.Time
}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	AddPeer    chan *Peer
	DelPeer    chan *Peer
}

func (p *Peer) Send(b []byte) error {
	_, err := p.conn.Write(b)

	if err != nil {
		return err
	}

	// response := make([]byte, 1024) // Adjust the buffer size as needed
	// n, err := p.conn.Read(response)

	// if err != nil {
	// 	return err
	// }
	// log.Print(string(response[:n]))

	return nil

}

// func (p *Peer) SendPing(b []byte) (string, error) {
// 	_, err := p.conn.Write(b)

// 	if err != nil {
// 		return "", err
// 	}

// 	response := make([]byte, 1024) // Adjust the buffer size as needed
// 	n, err := p.conn.Read(response)

// 	if err != nil {
// 		return "", err
// 	}

// 	return string(response[:n]), nil

// }

func (p *Peer) Rec(b []byte) error {
	_, err := p.conn.Read(b)
	return err
}

/*
	this function continuously reads data from the network connection (p.conn), creates a Message with the received data, and sends it to the msgch channel.

The reading process runs in an infinite loop until an error occurs, at which point the connection is closed.
*/
func (p *Peer) readLoop(msgch chan *Message) {
	// buf := make([]byte, 1024)
	for {

		msg := new(Message)

		if err := gob.NewDecoder(p.conn).Decode(msg); err != nil {
			logrus.Errorf("decode message error")
			break
		}

		msgch <- msg

	}

	
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

/* The purpose of this method is to continuously listen for incoming TCP connections on
the specified address and handle the accepted connections by creating
 Peer objects and passing them to the AddPeer channel for further processing.*/

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Error(err)
			continue
		}

		peer := &Peer{
			conn: conn,
		}

		t.AddPeer <- peer
	}

	return fmt.Errorf("TCP transport stopped reason")

}

package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	_ "reflect"
	"time"

	"net"
	"sync"

	"github.com/sirupsen/logrus"
)



type ServerConfig struct{
	ListenAddr string 
	Vesrison string
	
}


type Server struct {
	ServerConfig

	
	listener net.Listener
	transport *TCPTransport
	mu  sync.RWMutex
	peers map [net.Addr]*Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	
	
}




func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		
		ServerConfig: cfg,
		peers: make(map[net.Addr]*Peer),
		addPeer: make(chan *Peer, 500),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message),
	}
	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	tr.AddPeer = s.addPeer
	tr.DelPeer = s.addPeer
	

	return s
}

func (s *Server) Start() {
	go s.loop()
	fmt.Printf("server running on port %s \n", s.ListenAddr)
	logrus.WithFields(logrus.Fields{
		"port" : s.ListenAddr,
	}).Info("started new peer server")


	s.transport.ListenAndAccept()
	
}

func (s *Server) SendHandshake(p *Peer) error{
	hs := &Handshake {
		Version: "two",
		ListenAddr: s.ListenAddr,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {

	}

	return p.Send(buf.Bytes())

}
// handshake round trip
func (s *Server) isInPeerList(addr string) bool {
	for _, peer := range s.peers{
		if peer.listenAddr == addr {
			return true
		}
	}
	return false
}

// TODO 
func (s *Server) Connect(addr string) error {
	if s.isInPeerList(addr){
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	peer := &Peer {
			conn: conn,
		}

	s.addPeer <- peer
	// peer.Send([] byte("Vesrison 1"))

	return nil
}

/* The acceptLoop() method provided in the code is a method of the Server struct in Go. 
It implements an infinite loop that continuously accepts incoming TCP connections from the listener created in the Listen() method. 
For each accepted connection, it creates a new *Peer instance, adds it to the s.addPeer channel,
 and then starts handling the connection concurrently by calling s.handleConn(conn) in a Goroutine.*/


/* The Listen() method provided in the code is a method of the Server struct in Go. 
It handles the process of listening on a specified network address for incoming TCP connections.*/



/* The purpose of this loop() function is to continuously listen for incoming  *Peer objects on the 
s.addPeer channel and handle them as they arrive. When a new *Peer arrives, 
it adds the *Peer to the peers map with its remote 
network address as the key and then prints a message indicating the new peer connection.*/

func (s *Server) loop() {
	for {
		select{
		case peer := <- s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr" : peer.conn.RemoteAddr(),
			}).Info("peer disconneted")
			delete(s.peers, peer.conn.RemoteAddr())
		case peer:= <- s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				logrus.Errorf("handle peer error: %s", err)
			}
			
		case msg := <- s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}

type Handshake struct {
	Version string
	ListenAddr string
	
}

type PeerList struct {
	Peers []string
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From: from,
		Payload: payload,
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	s.SendHandshake(peer)
	hs, err := s.handShake(peer)
	if err != nil {
		peer.conn.Close()
		delete(s.peers, peer.conn.RemoteAddr())
		return fmt.Errorf("%s handshake with incoming peer failed: %s",s.ListenAddr, err)
		}

	go peer.readLoop(s.msgCh)

	logrus.WithFields(logrus.Fields{
		"addr" : peer.conn.RemoteAddr(),
	}).Info("connected")

	logrus.WithFields(logrus.Fields{
		"peer" : peer.conn.RemoteAddr(),
		"version": hs.Version,
		"listenAddr": peer.listenAddr,
		"we": s.ListenAddr,
	}).Info("handshake successfull: new peer connected")

	if err := s.sendPeerList(peer); err != nil {
		return fmt.Errorf("peerlist error : %s", err)
	}
	s.peers[peer.conn.RemoteAddr()] = peer	

	return nil
}

func (s *Server) sendPeerList(p *Peer ) error {
	
	peerList := PeerList{
		Peers : []string{},
	}

	

	for _, peer  := range s.peers{
		peerList.Peers = append(peerList.Peers, peer.listenAddr)
	}

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList)
	buf := new(bytes.Buffer)
	if err:= gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	return p.Send(buf.Bytes())

}

func (s* Server) handShake(p *Peer) (*Handshake, error) {

	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil, err
	}

	if s.Vesrison != hs.Version{
		return nil, fmt.Errorf("peer version not match %s", hs.Version)
	}

	p.listenAddr = hs.ListenAddr

	
	
	return hs, nil
}



func (s *Server ) handleMessage(msg *Message) error{

	
	switch v:=msg.Payload.(type){
	case PeerList:
		return s.handlePeerList(v)
	}
	// fmt.Print(msg)
	
	return nil
}

func (s *Server) handlePeerList(l PeerList) error {
	   
		logrus.WithFields(logrus.Fields{
		"we":s.ListenAddr,
		"list": l.Peers,

	}).Info("recieved message")
	
	for i :=0; i < len(l.Peers); i++ {
		fmt.Print(l.Peers)
		if err := s.Connect(l.Peers[i]); err != nil{
			logrus.Errorf("failed to connect peer: %s", err)
			continue
		}
	}
	return nil
}

func init () {
	gob.Register(PeerList{})

}




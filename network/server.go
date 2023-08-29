package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	_ "log"

	_ "reflect"

	"time"

	"net"
	"sync"

	"git.cs.bham.ac.uk/projects-2022-23/aaa234/metrics"
	"github.com/sirupsen/logrus"
)

const pingInterval = 5 * time.Second

type ServerConfig struct {
	ListenAddr string
	Vesrison   string
}

type Server struct {
	ServerConfig
	peerLock sync.RWMutex

	transport *TCPTransport

	peers        map[string]*Peer
	addPeer      chan *Peer
	delPeer      chan *Peer
	msgCh        chan *Message
	broadcastch  chan BroadcastTo
	broadcastch1 chan BroadcastTo
}

//server handles the communication, it handles the transport, it keeps track of peers
// A peer is a server on the other side of the connection.

type Handshake struct {
	Version    string
	ListenAddr string
}

type PeerList struct {
	Peers []string
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{

		ServerConfig: cfg,
		peers:        make(map[string]*Peer),
		addPeer:      make(chan *Peer, 20),
		delPeer:      make(chan *Peer),
		msgCh:        make(chan *Message, 500),
		broadcastch:  make(chan BroadcastTo, 500),
		broadcastch1: make(chan BroadcastTo, 500),
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
		"port": s.ListenAddr,
	}).Info("started new peer server")

	s.transport.ListenAndAccept()

}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		Version:    "two",
		ListenAddr: s.ListenAddr,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		log.Println("sendhand", err)
		return err

	}

	log.Println("bufff", buf)

	return p.Send(buf.Bytes())

}

// handshake round trip
func (s *Server) isInPeerList(addr string) bool {
	for _, peer := range s.peers {
		if peer.listenAddr == addr {
			return true
		}
	}
	return false
}

func (s *Server) Connect(addr string) error {

	if s.isInPeerList(addr) {
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	// conn1, err1 := net.DialTimeout("tcp", s.ListenAddr, 1*time.Second)
	if err != nil {
		return err
	}

	peer := &Peer{
		conn:   conn,
		status: true,
	}

	s.addPeer <- peer

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
		select {
		case msg := <-s.broadcastch:
			go func() {
				if err := s.Broadcast(msg); err != nil {

					logrus.Errorf("broadcast error: %s", err)
				}

			}()
		case msg1 := <-s.broadcastch1:
			go func() {
				if err := s.Broadcast1(msg1); err != nil {

					logrus.Errorf("broadcast error: %s", err)
				}

			}()
		case peer := <-s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("peer disconneted")
			delete(s.peers, peer.conn.RemoteAddr().String())
		case peer := <-s.addPeer:

			if err := s.handleNewPeer(peer); err != nil {

				logrus.Errorf("handle peer error: %s", err)
			}
		case msg := <-s.msgCh:
			go func() {

				// fmt.Print(msg)

				if err := s.handleMessage(msg); err != nil {
					logrus.Errorf("handle msg error: %s", err)
				}

			}()

		}
	}
}

func (s *Server) AddPeer(p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.listenAddr] = p
	p.status = true

}

// global variable for the  time
func (s *Server) Ping() {
	for {
		time.Sleep(pingInterval)
		for _, addr := range s.Peers() {
			logrus.WithFields(logrus.Fields{}).Info("PING" + addr)
			s.SendToPeers(recHeart, addr)

		}

	}

}

func (s *Server) Peers() []string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	peers := make([]string, len(s.peers))

	it := 0

	for _, peer := range s.peers {
		peers[it] = peer.listenAddr
		it++
	}

	return peers
}

func (s *Server) sendPeerList(p *Peer) error {

	peerList := PeerList{
		Peers: s.Peers(),
	}
	// for _, peer  := range s.peers{
	// 	peerList.Peers = append(peerList.Peers, peer.listenAddr)
	// }

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList, p.listenAddr)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	return p.Send(buf.Bytes())

}

func (s *Server) Broadcast(broadcastMsg BroadcastTo) error {

	msg := NewMessage(s.ListenAddr, broadcastMsg.Payload, broadcastMsg.To)

	// fmt.Print(msg)

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	peer, ok := s.peers[broadcastMsg.To]

	// fmt.Print(peer)

	//Buffering the message when nodes are not connected.

	if ok {
		// log.Print(peer.status, peer.listenAddr, "broad")

		go func(peer *Peer) {
			metric := metrics.NewMetrics(peer.conn.RemoteAddr().String())

			metric.FixWriteDuration()

			if err := peer.Send(buf.Bytes()); err != nil {
				logrus.Errorf("broadcast to peer error: %s", err)
			}

			if ShowLogs {
				logrus.WithFields(logrus.Fields{}).Info(metric.String())

			}
		}(peer)

		// if peer.status  {

		// } else {
		// 	logrus.Warnf("trying to ping disconnected peer %s", peer.listenAddr)
		// }

	} else {
		logrus.Warnf("trying to send to a disconnected peer %s", peer.listenAddr)
	}

	return nil
}

func (s *Server) Broadcast1(broadcastMsg BroadcastTo) error {

	msg := NewMessage(s.ListenAddr, broadcastMsg.Payload, broadcastMsg.To)

	// fmt.Print(msg)

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	peer, ok := s.peers[broadcastMsg.To]

	// fmt.Print(peer)

	//Buffering the message when nodes are not connected.

	if ok {
		// log.Print(peer.status, peer.listenAddr, "broad")

		go func(peer *Peer) {
			metric := metrics.NewMetrics(peer.conn.RemoteAddr().String())

			metric.FixWriteDuration()

			if err := peer.Send(buf.Bytes()); err != nil {
				logrus.Errorf("broadcast to peer error: %s", err)
			}

			if ShowLogs {
				logrus.WithFields(logrus.Fields{}).Info(metric.String())

			}
		}(peer)

		// if peer.status  {

		// } else {
		// 	logrus.Warnf("trying to ping disconnected peer %s", peer.listenAddr)
		// }

	} else {
		logrus.Warnf("trying to send to a disconnected peer %s", peer.listenAddr)
	}

	return nil
}

func (s *Server) SendToPeers(payload any, addr string) {

	s.broadcastch <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

}

func (s *Server) SendToPeers1(payload any, addr string) {

	s.broadcastch1 <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

}

func (s *Server) handShake(p *Peer) (*Handshake, error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		// log.Println("hand", err)
		return nil, err
	}

	log.Print(hs)

	if s.Vesrison != hs.Version {
		return nil, fmt.Errorf("peer version not match %s", hs.Version)
	}

	p.listenAddr = hs.ListenAddr

	return hs, nil

}

func init() {
	gob.Register(PeerList{})

}

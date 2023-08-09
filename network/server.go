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
	"gitlab.com/adeola/messaging-library/metrics"
)



type ServerConfig struct{
	ListenAddr string 
	Vesrison string
	APIlistenAddr string
	
}


type Server struct {
	ServerConfig
	peerLock sync.RWMutex

	
	listener net.Listener
	transport *TCPTransport
	mu  sync.RWMutex
	peers map [string]*Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	broadcastch chan BroadcastTo
	
	
	
}


//server handles the communication, it handles the transport, it keeps track of peers
// A peer is a server on the other side of the connection.

type Handshake struct {
	Version string
	ListenAddr string
	
}

type PeerList struct {
	Peers []string
}


func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		
		ServerConfig: cfg,
		peers: make(map[string]*Peer),
		addPeer: make(chan *Peer, 20),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message, 100),
		broadcastch: make(chan BroadcastTo, 100),
		
		
	}

	// s.metrics := newMetrics(cfg.ListenAddr)
	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	tr.AddPeer = s.addPeer
	tr.DelPeer = s.addPeer



	// go func(s *Server){
	// 	apiServer := NewAPIServer(cfg.APIlistenAddr)
		
	// 	logrus.WithFields(logrus.Fields{
	// 		"listenAddr": cfg.APIlistenAddr,
	// 	}).Info("starting API server")
	// 	apiServer.Run()

	// }(s)
	

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
		return err

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
	// conn1, err1 := net.DialTimeout("tcp", s.ListenAddr, 1*time.Second)
	if err != nil {
		return err
	}

	// if err1 != nil {
	// 	return err
	// }
	peer := &Peer {
			conn: conn,
		}

	// peer1 := &Peer {
	// 		conn: conn1,
	// 	}
	// fmt.Print(peer1)

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
		case msg := <- s.broadcastch:
			
			go func(){
				if err := s.Broadcast(msg); err != nil {
					logrus.Errorf("broadcast error: %s", err)
				}
				
			}()
		case peer := <- s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr" : peer.conn.RemoteAddr(),
			}).Info("peer disconneted")
			delete(s.peers, peer.conn.RemoteAddr().String())
		case peer:= <- s.addPeer:
			
			if err := s.handleNewPeer(peer); err != nil {
				
				logrus.Errorf("handle peer error: %s", err)
			}
			
		case msg := <- s.msgCh:	
			go func(){

				// fmt.Print(msg)
				
				if err := s.handleMessage(msg); err != nil {
					logrus.Errorf("handle msg error: %s", err)
				}

			}()
				
		}
	}
}





func (s *Server) AddPeer (p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.listenAddr] = p

}

func (s *Server) Ping () {
	interval := 2 * time.Second
	for {
		// if len(s.Peers()) == 0 {
		// 	fmt.Print(0)
		// }

		time.Sleep(interval)
		for _, addr := range s.Peers(){
		logrus.WithFields(logrus.Fields{
	     }).Info("Sending PING")
		s.SendToPeers("PING", addr)
		
		
	}

	}
	
}



func (s *Server) Peers() [] string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	peers := make([] string, len(s.peers))

	it := 0

	for _, peer := range s.peers {
		peers[it] = peer.listenAddr
		it ++
	}

	return peers
}








func (s *Server) sendPeerList(p *Peer ) error {
	
	peerList := PeerList{
		Peers : s.Peers(),
	}
	// for _, peer  := range s.peers{
	// 	peerList.Peers = append(peerList.Peers, peer.listenAddr)
	// }

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList, p.listenAddr)
	buf := new(bytes.Buffer)
	if err:= gob.NewEncoder(buf).Encode(msg); err != nil {
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
			
			go func(peer *Peer) {

				metric := metrics.NewMetrics(peer.conn.RemoteAddr().String())

				metric.FixWriteDuration()

				logrus.WithFields(logrus.Fields{

				}).Info(metric.String())

				if err := peer.Send(buf.Bytes()); err != nil {
					logrus.Errorf("broadcast to peer error: %s", err)
				}
				
			}(peer)
			
		}

	return nil
}



func (s *Server) SendToPeers(payload any, addr string) {
	
	 s.broadcastch <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	
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





func init () {
	gob.Register(PeerList{})

}




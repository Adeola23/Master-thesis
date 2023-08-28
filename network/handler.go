package network

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/adeola/messaging-library/metrics"
)

const (
	retryCount         = 3
	waitTime           = 2 * time.Second
	recHeart    string = "PING"
	resHeart    string = "PONG"
	SendMessage string = "HEY SERVER"
	ResMessage  string = "HI SERVER WHATS UP"
	defaultRes  string = "Unkown message"
	nilResponse string = "NIL MESSAGE SENT"
)

var TopologyInstance Topology

var ShowLogs = true

func (s *Server) handleNewPeer(peer *Peer) error {

	s.SendHandshake(peer)

	hs, err := s.handShake(peer)
	if err != nil {
		peer.conn.Close()
		delete(s.peers, peer.conn.RemoteAddr().String())
		return fmt.Errorf("%s handshake with incoming peer failed: %s", s.ListenAddr, err)
	}
	metric := metrics.NewMetrics(peer.conn.RemoteAddr().String())
	go peer.readLoop(s.msgCh)

	if ShowLogs {
		logrus.WithFields(logrus.Fields{
			"addr": peer.conn.RemoteAddr(),
		}).Info("connected")

		logrus.WithFields(logrus.Fields{
			"peer":       peer.conn.RemoteAddr(),
			"version":    hs.Version,
			"listenAddr": peer.listenAddr,
			"we":         s.ListenAddr,
		}).Info("handshake successfull: new peer connected")

	}

	if TopologyInstance.Structured {
		if err := s.sendPeerList(peer); err != nil {
			return fmt.Errorf("peerlist error : %s", err)
		}

	}

	// s.peers[peer.conn.RemoteAddr()] = peer

	s.AddPeer(peer)

	metric.FixHandshake()

	if ShowLogs {
		logrus.WithFields(logrus.Fields{}).Info(metric.String())

	}

	return nil
}

func (s *Server) handlePeerList(l PeerList) error {

	// 	logrus.WithFields(logrus.Fields{
	// 	"we":s.ListenAddr,
	// 	"list": l.Peers,

	// }).Info("recieved message")

	for i := 0; i < len(l.Peers); i++ {

		if err := s.Connect(l.Peers[i]); err != nil {
			logrus.Errorf("failed to connect peer: %s", err)
			continue
		}
	}
	return nil
}

func (s *Server) resp(msg any, addr string) {
	s.broadcastch <- BroadcastTo{
		To:      addr,
		Payload: msg,
	}

}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {

	case PeerList:
		return s.handlePeerList(v)
	case string:

		if msg.Payload == recHeart || msg.Payload == resHeart {
			return s.handleHearbeat(msg.Payload, msg.From, msg.To)
		} else if msg.Payload == SendMessage || msg.Payload == ResMessage {
			return s.handleMsg(msg.Payload, msg.From, msg.To)

		} else if msg.Payload == nilResponse {
			return s.handleMsg(msg.Payload, msg.From, msg.To)

		} else {
			log.Println(msg.Payload)
			return s.handleUnkown(msg.Payload, msg.From, msg.To)

		}

	case nil:
		return s.handleMsg(msg.Payload, msg.From, msg.To)

	case int:
	}

	return nil
}

func (s *Server) handleMsg(msg any, from string, to string) error {
	metric := metrics.NewMetrics(from)

	metric.FixReadDuration()

	if msg == "" || msg == nil {
		// Handle empty string or nil message
		s.resp(nilResponse, from)
		logrus.Errorf(nilResponse)
	} else if msg == SendMessage {
		s.resp(ResMessage, from)

	} else if msg != ResMessage {
		s.resp(defaultRes, from)
		log.Println(msg)

	}

	if ShowLogs {
		logrus.WithFields(logrus.Fields{
			"sender":   from,
			"message":  msg,
			"receiver": to,
		}).Info(metric.String())

	}

	return nil

}

func (s *Server) handleUnkown(msg any, from string, to string) error {
	metric := metrics.NewMetrics(from)

	if msg == defaultRes {
		if ShowLogs {
			logrus.WithFields(logrus.Fields{
				"sender":   from,
				"message":  msg,
				"receiver": to,
			}).Info(metric.String())

		}

	} else {
		metric.FixReadDuration()
		time.Sleep(waitTime)
		for i := 0; i < retryCount; i++ {
			s.resp(defaultRes, from)
			logrus.Warnf("Unknow message is sent from %s to %s", from, to)
		}

		if ShowLogs {
			logrus.WithFields(logrus.Fields{
				"sender":   from,
				"message":  msg,
				"receiver": to,
			}).Info(metric.String())

		}

	}

	return nil

}

func (s *Server) handleHearbeat(msg any, from string, to string) error {
	metric := metrics.NewMetrics(from)

	metric.FixReadDuration()

	recMsg := msg

	switch recMsg {
	case recHeart:
		s.SendToPeers1(resHeart, from)
		logrus.WithFields(logrus.Fields{}).Info("PONG" + from)
		if ShowLogs {
			logrus.WithFields(logrus.Fields{
				"sender":   from,
				"message":  recHeart,
				"receiver": to,
			}).Info(metric.String())

		}
	case resHeart:

		if ShowLogs {
			logrus.WithFields(logrus.Fields{
				"status": "Alive",
				"Node":   to,
			}).Info(metric.String())

		}
		s.UpdatePeerStatus(from, true)

	}
	return nil
}

func (s *Server) UpdatePeerStatus(addr string, status bool) {

	if peer, ok := s.peers[addr]; ok {
		peer.LastPingTime = time.Now()
		peer.status = status

	}

}

func (s *Server) IsPeerResponsive() {

	for _, peer := range s.peers {
		log.Println("After update check:", peer.LastPingTime)
		elapsedTime := time.Since(peer.LastPingTime)
		comparisonResult := elapsedTime <= pingInterval*2
		if comparisonResult {
			peer.status = true
			log.Println(peer.status, peer.listenAddr, peer.LastPingTime, "compa")
		} else {
			peer.status = false

		}
		log.Print(comparisonResult, peer.listenAddr, "check")
		if !peer.status {
			logrus.WithFields(logrus.Fields{
				"source": s.ListenAddr,
			}).Warn("Peer " + peer.listenAddr + " is unresponsive")

			time.Sleep(time.Second * 10)
			for i := 0; i < retryCount; i++ {

				delete(s.peers, peer.listenAddr)

				err := s.Connect(peer.listenAddr)

				if err != nil {
					logrus.Error(err)
				}

				if err == nil {
					//success
					// s.UpdatePeerStatus(peer.listenAddr)
					peer.LastPingTime = time.Now()
					peer.status = true
					log.Println("CONNECTED")
					break
				}

				if i < retryCount-1 {
					time.Sleep(waitTime)
					fmt.Printf("Retrying in %s...\n", waitTime)

				}

			}
			// if comparisonResult {
			// 	break
			// 	panic("jsbdjsb")

			// }

		}

	}

}

func (s *Server) PingPeer(addr string) error {

	if peer, ok := s.peers[addr]; ok {
		pingMessage := "PING"
		_, err := peer.conn.Write([]byte(pingMessage))
		if err != nil {
			return err
		}

		// Set a timeout for waiting for a ping response
		responseDeadline := time.Now().Add(2 * time.Second)
		peer.conn.SetReadDeadline(responseDeadline)

		// Attempt to read the response
		response := make([]byte, len(pingMessage))
		_, err = peer.conn.Read(response)
		if err != nil {
			// Handle the error, such as connection timeout or closed connection
			return err
		}

		// Response received, update the peer's status
		peer.status = true

	}
	// Send a ping to the peer

	return nil
}

func (s *Server) ReconnectPeer(peer *Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	//  err := peer.conn.Close() // Close the existing connection

	// log.Println(err, "errrrr")
	// if err != nil {
	// 	return err
	// }

	// Attempt to establish a new connection
	newConn, err := net.DialTimeout("tcp", peer.listenAddr, 1*time.Second)
	if err != nil {
		return err
	}

	// Update the peer's connection status and connection object
	peer.conn = newConn
	peer.status = true

	log.Println(newConn)

	// Perform any necessary post-connection setup

	return nil
}

func (s *Server) Disconnect(addr string) error {
	peer, ok := s.peers[addr]
	if !ok {
		return fmt.Errorf("%v failed to disconnect: unknown peer: %v", s.ListenAddr, addr)
	}

	if ok {
		err := peer.conn.Close()

		log.Println("peer list before >>>>", s.peers[addr])

		delete(s.peers, peer.listenAddr)

		log.Println("peer list after >>>>", s.peers[addr])

		if err != nil {
			return fmt.Errorf("%v failed to disconnect: %v", addr, err)

		} else {

			peer.conn = nil

			log.Println("Connection closed successfully")
		}
	}

	return nil
}

func (s *Server) StartPeerStatusChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.IsPeerResponsive()
		}
	}
}

// func (s *Server) checkPeerStatus() {
// 	for _, peer := range s.peers {

// 		if !peer.connected {
// 			// peer.conn.Close()
// 			delete(s.peers, peer.conn.RemoteAddr().String())
// 			logrus.Errorf("peer %s disconnected and deleted from : %s", peer.listenAddr, s.ListenAddr)

// 			for i := 0; i < retryCount; i++ {

// 				err := s.Connect(peer.listenAddr)

// 				if err != nil {
// 					logrus.Error(err)
// 				}

// 				if err == nil {
// 					//success
// 					// s.UpdatePeerStatus(peer.listenAddr)
// 					log.Println("CONNECTED")
// 					break
// 				}

// 				if i < retryCount-1 {
// 					time.Sleep(waitTime)
// 					fmt.Printf("Retrying in %s...\n", waitTime)

// 				}

// 			}

// 		}

// 	}
// }

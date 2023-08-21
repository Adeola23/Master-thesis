package network

import (
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/adeola/messaging-library/metrics"
)

const (
	retryCount         = 3
	waitTime           = 10 * time.Second
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
			// return s.handleUnkown(msg.Payload, msg.From, msg.To)

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

	// if recMsg != SendMessage {
	// 	s.resp("")

	// }
	log.Print(msg)

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

	// switch recMsg {

	// case SendMessage:

	// case defaultRes:
	// 	s.resp(ResMessage, from)
	// default:

	// }

	if ShowLogs {
		logrus.WithFields(logrus.Fields{
			"sender":   from,
			"message":  msg,
			"receiver": to,
		}).Info(metric.String())

	}

	return nil

}
// func (s *Server) handleUnkown(msg any, from string, to string) error {
// 	metric := metrics.NewMetrics(from)

// 	metric.FixReadDuration()
// 	time.Sleep(waitTime)
// 	for i := 0; i < retryCount; i++ {
// 		s.resp(defaultRes, from)
// 	}

// 	if ShowLogs {
// 		logrus.WithFields(logrus.Fields{
// 			"sender":   from,
// 			"message":  msg,
// 			"receiver": to,
// 		}).Info(metric.String())

// 	}

// 	return nil

// }

func (s *Server) handleHearbeat(msg any, from string, to string) error {
	metric := metrics.NewMetrics(from)

	metric.FixReadDuration()

	recMsg := msg

	switch recMsg {
	case recHeart:
		s.resp(resHeart, from)
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
				"status":  "Alive",
				"Node": to,
			}).Info(metric.String())

		}
		s.UpdatePeerStatus(from, true)
	case nil:
		s.UpdatePeerStatus(from, false)
	}

	return nil

}

func (s *Server) UpdatePeerStatus(addr string, connected bool) {
	if peer, ok := s.peers[addr]; ok {
		peer.connected = connected
	}

}

func (s *Server) StartPeerStatusChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkPeerStatus()
		}
	}
}

func (s *Server) checkPeerStatus() {
	for _, peer := range s.peers {

		if !peer.connected {
			// peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr().String())
			logrus.Errorf("peer %s disconnected and deleted from : %s", peer.listenAddr, s.ListenAddr)

			for i := 0; i < retryCount; i++ {

				err := s.Connect(peer.listenAddr)

				if err != nil {
					logrus.Error(err)
				}

				if err == nil {
					//success
					s.UpdatePeerStatus(peer.listenAddr, true)
					log.Println("CONNECTED")
					break
				}

				if i < retryCount-1 {
					time.Sleep(waitTime)
					fmt.Printf("Retrying in %s...\n", waitTime)

				}

			}

		}

	}
}

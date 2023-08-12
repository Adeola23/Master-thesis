package network

import (
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.com/adeola/messaging-library/metrics"
)

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

	logrus.WithFields(logrus.Fields{
		"addr": peer.conn.RemoteAddr(),
	}).Info("connected")

	logrus.WithFields(logrus.Fields{
		"peer":       peer.conn.RemoteAddr(),
		"version":    hs.Version,
		"listenAddr": peer.listenAddr,
		"we":         s.ListenAddr,
	}).Info("handshake successfull: new peer connected")

	if err := s.sendPeerList(peer); err != nil {
		return fmt.Errorf("peerlist error : %s", err)
	}
	// s.peers[peer.conn.RemoteAddr()] = peer

	s.AddPeer(peer)

	metric.FixHandshake()

	logrus.WithFields(logrus.Fields{}).Info(metric.String())

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

func (s *Server) handleMsg(msg any, from string, to string) error {
	metric := metrics.NewMetrics(from)

	metric.FixReadDuration()

	recMsg := msg

	switch recMsg {
	case "PING":
		s.resp("PONG", from)
	case "PONG":
		s.UpdatePeerStatus(from, true)
	case nil:
		s.UpdatePeerStatus(from, false)
	}
	logrus.WithFields(logrus.Fields{
		"sender":  from,
		"message": msg,
	}).Info(metric.String())

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
	retryCount := 3
	waitTime := 10 * time.Second
	for _, peer := range s.peers {

		if !peer.connected {
			peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr().String())
			logrus.Errorf("peer %s disconnected and deleted from : %s", peer.listenAddr, s.ListenAddr)

			time.Sleep(waitTime)
			for i := 0; i < retryCount; i++ {
				log.Println("RETRY")
				err := s.Connect(peer.listenAddr)
				log.Println("RETRY22")

				if err != nil {
					logrus.Error(err)

				}

				if err == nil {

					s.UpdatePeerStatus(peer.listenAddr, true)

					//success
					break
				}

				if i < retryCount-1 {
					time.Sleep(waitTime)
					fmt.Printf("Retrying in %s...\n", waitTime)

				}

				peer.conn.Close()

			}

		}

	}
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {

	case PeerList:
		return s.handlePeerList(v)
	case string:
		return s.handleMsg(msg.Payload, msg.From, msg.To)
	case int:
	}

	return nil
}

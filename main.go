package main

import (
	_"fmt"
	"time"

	"gitlab.com/adeola/messaging-library/network"
)

func makeServerAndStart(addr string) *network.Server{
	cfg := network.ServerConfig{
		ListenAddr: addr,
		Vesrison: "two",
		
	}
	server := network.NewServer(cfg)
	time.Sleep(1 * time.Second)

	go server.Start()

	return server

}

func main () {
	peerA  := makeServerAndStart(":3000")
	peerB := makeServerAndStart(":4000")
	peerC := makeServerAndStart(":4300")
	peerD := makeServerAndStart(":4500")
	// peerE  := makeServerAndStart(":3000")
	// peerF := makeServerAndStart(":4200")
	// peerG := makeServerAndStart(":4100")
	// peerH := makeServerAndStart(":4150")
	// peerI  := makeServerAndStart(":3200")
	// peerJ := makeServerAndStart(":4400")
	// peerK := makeServerAndStart(":4600")
	// peerL := makeServerAndStart(":4800")
	


	time.Sleep(1 * time.Second)

	peerB.Connect(peerA.ListenAddr)

	time.Sleep(1 * time.Second)

	peerC.Connect(peerB.ListenAddr)

	time.Sleep(1 * time.Second)

	peerA.Connect(peerD.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerE.Connect(peerF.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerG.Connect(peerF.ListenAddr)

	

	



	select{}

}

// func testPersConnected(s *network.Server){
// 	peers := [] string {
// 		":3000",
// 		":3500",
// 		":3600",
// 		":4000",
// 		":4400",
// 		":4500",
// 		":4600",
// 	}

// 	p := s.Peers()

// 	for i := 0; i < len(peers); i++ {
// 		for x := 0; x < len(peers); x++ {

// 		}
// 	}
// }



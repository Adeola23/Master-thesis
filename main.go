package main

import (
	_"encoding/gob"
	_"fmt"
	_ "fmt"
	_"net"
	"time"

	_"github.com/sirupsen/logrus"
	"gitlab.com/adeola/messaging-library/network"
)

func makeServerAndStart(addr, apiddr string) *network.Server{
	cfg := network.ServerConfig{
		ListenAddr: addr,
		APIlistenAddr: apiddr,
		Vesrison: "two",
		
	}
	server := network.NewServer(cfg)
	time.Sleep(1 * time.Second)

	go server.Start()

	return server

}

func main () {
	peerA  := makeServerAndStart(":3000", ":3001")
	peerB := makeServerAndStart(":4000", ":4001")
	peerC := makeServerAndStart(":4300", ":4301")
	peerD := makeServerAndStart(":4500", ":4501")
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

	peerA.Connect(peerD.ListenAddr)



	// time.Sleep(1 * time.Second)

	// peerA.Connect(peerB.ListenAddr)

	// to := []string{":4000"}
	// to1 := []string{":4500"}

	// peerA.SendToPeers("Hey", to...)

	// time.Sleep(1 * time.Second)

	



	time.Sleep(1 * time.Second)

	peerC.Connect(peerB.ListenAddr)

	to := []string{":4000"}
	//to1 := []string{":4500"}

	time.Sleep(1 * time.Second)

	peerC.SendToPeers("Hey", to...)

	to1 := []string{":4500"}
	//to1 := []string{":4500"}

	time.Sleep(1 * time.Second)

	peerA.SendToPeers("HI 45000", to1...)


	to2 := []string{":3000"}
	//to1 := []string{":4500"}

	time.Sleep(1 * time.Second)

	peerC.SendToPeers("HI 3000", to2...)


	to3 := []string{":4300"}
	//to1 := []string{":4500"}

	time.Sleep(1 * time.Second)

	peerB.SendToPeers("HI 4300", to3...)


	to4 := []string{":4000"}
	//to1 := []string{":4500"}

	time.Sleep(1 * time.Second)

	peerA.SendToPeers("HI 4000", to4...)



	

	// time.Sleep(1 * time.Second)

	// peerA.SendToPeers("Yo", to1...)

	// peerE.Connect(peerF.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerG.Connect(peerF.ListenAddr)

	// msg := new(network.Message)


	//  time.Sleep(1 * time.Second)
	// conn, err := net.DialTimeout("tcp", ":3000", 1*time.Second)
	// if err != nil {
	// 	fmt.Print(err)
	// }

	// time.Sleep(1 * time.Second)
    // if err := gob.NewDecoder(conn).Decode(msg); err != nil {
	// 		fmt.Print(err)
			
	// }
	

	

	



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



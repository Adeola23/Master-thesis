package main

import (
	_ "encoding/gob"
	_ "fmt"
	"log"

	_ "log"
	_ "net"

	"time"

	"git.cs.bham.ac.uk/projects-2022-23/aaa234/network"
	_ "github.com/sirupsen/logrus"
)

func makeServerAndStart(addr string) *network.Server {
	cfg := network.ServerConfig{
		ListenAddr: addr,
		Vesrison:   "two",
	}
	server := network.NewServer(cfg)
	time.Sleep(1 * time.Second)

	go server.Start()

	return server

}

func main() {
	network.TopologyInstance = network.InitializeTopology(true)
	network.ShowLogs = false

	peerA := makeServerAndStart(":3000")
	peerB := makeServerAndStart(":4000")
	peerC := makeServerAndStart(":4300")
	// peerD := makeServerAndStart(":4500")
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

	peerA.Connect(peerC.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerD.Connect(peerE.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerF.Connect(peerG.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerH.Connect(peerI.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerJ.Connect(peerA.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerK.Connect(peerJ.ListenAddr)

	// time.Sleep(1 * time.Second)

	// peerL.Connect(peerC.ListenAddr)

	// handles 9000 for a period and further breaks connection
	// messageRate := 1000 // messages per second
	// limiter := time.Tick(time.Second / time.Duration(messageRate))

	// for {
	// 	select {
	// 	case <-limiter:
	// 		peerC.SendToPeers(network.SendMessage, ":4000")

	// 	}
	// }

	// totalMessages := 2

	// // Establish connection and initialize sender
	time.Sleep(1 * time.Second)
	peerC.SendToPeers("HEYYY", ":3000")

	// start := time.Now()

	// for i := 0; i < totalMessages; i++ {

	// 	// Send message of messageSize
	// }
	// elapsed := time.Since(start)

	// throughput := float64(totalMessages) / elapsed.Seconds()
	// fmt.Printf("Throughput: %.2f messages/second\n", throughput)

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

	go peerA.Ping()
	// go peerB.Ping()
	// go peerC.Ping()
	// go peerD.Ping()

	time.Sleep(5 * time.Second)
	// go peerC.StartPeerStatusChecker(time.Second * 5)
	go peerB.StartPeerStatusChecker(time.Second * 5)
	go peerA.StartPeerStatusChecker(time.Second * 5)
	// log.Print("DONE")
	//go peerB.StartPeerStatusChecker(time.Second * 3)

	// time.Sleep(time.Second * 7)
	// peerA.Disconnect(":4000")
	// peerA.Disconnect(":4300")

	// log.Println("DISCONNNECTED")

	time.Sleep(7 * time.Second)
	peerA.UpdatePeerStatus(":4000", false)

	// time.Sleep(15 * time.Second)

	// peerA.Connect(peerB.ListenAddr)

	log.Println("DISCONNNECTED")

	time.Sleep(20 * time.Second)
	peerB.SendToPeers(network.SendMessage, peerA.ListenAddr)

	// Simulate a peer becoming reconnected

	select {}

}

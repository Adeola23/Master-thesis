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
	


	time.Sleep(1 * time.Second)

	peerB.Connect(peerA.ListenAddr)

	time.Sleep(1 * time.Second)

	peerC.Connect(peerB.ListenAddr)

	time.Sleep(1 * time.Second)

	peerA.Connect(peerD.ListenAddr)



	select{}

}



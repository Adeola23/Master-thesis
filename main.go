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


	time.Sleep(1 * time.Second)

	peerB.Connect(peerA.ListenAddr)


	select{}

}



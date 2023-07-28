package main

import (
	"fmt"
	"time"

	"gitlab.com/adeola/messaging-library/network"
)


func main () {


	cfg := network.ServerConfig{
		ListenAddr: ":3000",
		Vesrison: "two",
		
	}
	server := network.NewServer(cfg)
	go server.Start()

	time.Sleep(1 * time.Second)

	remoteCfg := network.ServerConfig{
		ListenAddr: ":4000",
		Vesrison: "two",
		
	}
	remoteServer := network.NewServer(remoteCfg)
	go remoteServer.Start()
	connectTo := ":3000"
	infoTo := fmt.Sprintf("SERVER %s CONNECTED TO SERVER %s", remoteServer.ListenAddr, connectTo)
	time.Sleep(1 * time.Second)
	remoteServer.Connect(connectTo, infoTo)
	

    
	remoteCfg1 := network.ServerConfig{
		ListenAddr: ":5000",
		Vesrison: "two",
		
	}
	remoteServer1 := network.NewServer(remoteCfg1)
	connectTo1 := ":4000"
	infoTo1 := fmt.Sprintf("SERVER %s CONNECTED TO SERVER %s", remoteServer1.ListenAddr, connectTo1)
	go remoteServer1.Start()
	time.Sleep(1 * time.Second)
	remoteServer1.Connect(connectTo1, infoTo1)

	
	remoteCfg2 := network.ServerConfig{
		ListenAddr: ":6000",
		Vesrison: "two",
		
	}
	remoteServer2 := network.NewServer(remoteCfg2)
	connectTo2 := ":5000"
	infoTo2 := fmt.Sprintf("SERVER %s CONNECTED TO SERVER %s", remoteServer2.ListenAddr, connectTo2)
	go remoteServer2.Start()
	time.Sleep(1 * time.Second)
	remoteServer2.Connect(connectTo2, infoTo2)

	select{}

}



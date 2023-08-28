## Messaging Sytem

This project implements a lightweight and configurable messaging system in Golang, designed for easy integration into distributed applications.

### Usage 

```bash
func main() {
    //Select topology (Structured and unstructured)
    network.TopologyInstance = network.InitializeTopology(true)

    //Turn on or turn off logs
	network.ShowLogs = false

    //Initialize servers
	peerA := makeServerAndStart(":3000")
	peerB := makeServerAndStart(":4000")

    //Contect the peers
    peerA.Connect(peerB.ListenAddr) or peerB.Connect(peerA.ListenAddr)

	// Send a message (network.SendMessage is configurable)
	peerB.SendToPeers(network.SendMessage, peerA.ListenAddr)
    

    //Periodic pinging(pingInterval is configurable)
    go peerA.Ping()
	go peerB.Ping()

    //Check peer status
    go peerB.StartPeerStatusChecker(time.Second * 5)
	go peerA.StartPeerStatusChecker(time.Second * 5)
}
```

## Functionalities

-- Gossip protocol: Each node maintains a list of other nodes it's aware of and share such list with connected peers.

-- Messaging: Send and receive messages between nodes in the network.

-- Heartbeat: Keep track of node availability with heartbeat signals.

-- TCP Protocol: Communicate using the TCP protocol for reliability.

-- Handshake: Establish connections with a handshake mechanism.

-- Logging: turn logs on/off.

-- Message Serialization/Deserialization: Messages are automatically serialized and deserialized.


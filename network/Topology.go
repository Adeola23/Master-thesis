package network

type Topology struct {
	Structured bool
}

func InitializeTopology(structured bool) Topology {
	return Topology{
		Structured: structured,
	}
}
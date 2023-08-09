package network

type Message struct {
	Payload any
	From    string
	To 		string
}

type BroadcastTo struct {
	To      string
	Payload any
}

type MessageReady struct{}

type MessageState struct{
	listenAddr string

	broadcastch chan BroadcastTo

}

func NewMessage(from string, payload any, to string) *Message {
	return &Message{
		From:    from,
		Payload: payload,
		To: to,
	}
}

func (m *MessageState) sendToPlayers(payload any, addr string) {
	m.broadcastch <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}
}
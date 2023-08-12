package network

type Message struct {
	Payload any
	From    string
	To      string
}

type BroadcastTo struct {
	To      string
	Payload any
}

func NewMessage(from string, payload any, to string) *Message {
	return &Message{
		From:    from,
		Payload: payload,
		To:      to,
	}
}

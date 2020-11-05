package net

type Message struct {
	data interface{}
	cmd  uint16
	Next bool
}

func NewMessage(data interface{}, next ...bool) *Message {
	msg := &Message{
		data: data,
	}
	if len(next) > 0 && next[0] {
		msg.Next = true
	}
	return msg
}

func (this *Message) GetData() interface{} {
	return this.data
}

func (this *Message) GetCmd() uint16 {
	return this.cmd
}

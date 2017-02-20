package rmq

import (
	"github.com/streadway/amqp"
)

type Message struct {
	delivery *amqp.Delivery
}

func (m *Message) ack(success, multiple bool) error {
	if success {
		return m.delivery.Ack(multiple)
	} else {
		return m.delivery.Nack(multiple, false)
	}
}

func (m *Message) Ack(success bool) error {
	return m.ack(success, false)
}

func (m *Message) AckMultiple(success bool) error {
	return m.ack(success, true)
}

func (m *Message) Body() []byte {
	return m.delivery.Body
}

func (m *Message) Headers() map[string]interface{} {
	return map[string]interface{}(m.delivery.Headers)
}

func (m *Message) Reject(requeue bool) error {
	return m.delivery.Reject(requeue)
}

func newMessage(d *amqp.Delivery) *Message {
	return &Message{
		delivery: d,
	}
}

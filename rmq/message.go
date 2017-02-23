package rmq

import (
	"github.com/streadway/amqp"
)

type Message struct {
	delivery    amqp.Delivery
	Body        []byte
	Headers     map[string]interface{}
}

func newMessage(d *amqp.Delivery) *Message {
	return &Message{
		delivery: *d,
		Body: d.Body,
		Headers: map[string]interface{}(d.Headers),
	}
}

func (m *Message) ack(success, multiple bool) error {
	//m.delivery.DeliveryTag = m.DeliveryTag
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

func (m *Message) Reject(requeue bool) error {
	return m.delivery.Reject(requeue)
}

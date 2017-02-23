package rmq

import (
	"fmt"
	"github.com/streadway/amqp"
	//"log"
)

type Channel struct {
	channel *amqp.Channel
}

func newChannel(r *RMQ, prefetchCount int) (*Channel, error) {
	var err error

	c := &Channel{}
	c.channel, err = r.connection.Channel()
	if err != nil {
		return nil, err
	}

	c.channel.Qos(prefetchCount, 0, false)
	return c, nil
}

func (c *Channel) BasicConsume(queue string, ctag string, callback ConsumerCallback) error {
	deliveries, err := c.channel.Consume(
		queue, // name
		ctag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	go func() {
		for d := range deliveries {
			msg := newMessage(&d)
			go callback(msg)
		}
	}()

	return nil
}

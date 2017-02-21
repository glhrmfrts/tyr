package rmq

import (
	"fmt"
	"github.com/streadway/amqp"
)

type Channel struct {
	channel *amqp.Channel
}

type ConsumerCallback func(msg *Message)

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
			go callback(newMessage(&d))
		}
	}()

	return nil
}

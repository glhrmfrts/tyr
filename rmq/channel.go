package rmq

import (
	"fmt"

	"github.com/streadway/amqp"
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
		ctag,  // consumerTag,
		false, // noAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
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

func (c *Channel) BasicPublish(exchange string, rk string, body string, headers map[string]interface{}) error {
	err := c.channel.Publish(
		exchange,
		rk,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:      headers,
			Body:         []byte(body),
			DeliveryMode: amqp.Transient,
		},
	)
	if err != nil {
		return fmt.Errorf("Basic publish: %s", err)
	}

	return nil
}

func (c *Channel) QueueBind(name, key, exchange string, args map[string]interface{}) error {
	return c.channel.QueueBind(name, key, exchange, false, amqp.Table(args))
}

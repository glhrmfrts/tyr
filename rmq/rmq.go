package rmq

import (
	"fmt"
	"log"
	"github.com/streadway/amqp"
)

type Channel = amqp.Channel
type Return = amqp.Return

type RMQ struct {
	channels map[string]*Channel
	connection *amqp.Connection
}

func (r *RMQ) Channel(name string, prefetchCount int) (*Channel, error) {
	r.assertChannels()

	if r.connection == nil {
		return nil, fmt.Errorf("No connection")
	}

	if _, ok := r.channels[name]; !ok {
		channel, err := newChannel(r, prefetchCount)
		if err != nil {
			return nil, err
		}

		ch := make(chan *amqp.Error)
		channel.NotifyClose(ch)

		go func() {
			for err := range ch {
				log.Println(err)
			}
		}()

		r.channels[name] = channel
	}

	return r.channels[name], nil
}

func (r *RMQ) announce(channel *Channel, sett *Settings) error {
	var err error

	err = channel.ExchangeDeclare(
		sett.Exchange.Topic,
		"topic",
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	err = channel.ExchangeDeclare(
		sett.Exchange.Headers,
		"headers",
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	for _, s := range sett.Services {
		_, err = channel.QueueDeclare(
			s.Queue,
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // noWait
			nil,       // arguments
		)
		if err != nil {
			return fmt.Errorf("Queue Declare: %s", err)
		}

		err = channel.QueueBind(
			s.Queue,
			s.RoutingKey,
			sett.Exchange.Topic,
			false,      // noWait
			nil,        // arguments
		)
		if err != nil {
			return fmt.Errorf("Queue Bind: %s", err)
		} else {
			log.Printf("RK %s -> Queue %s", s.RoutingKey, s.Queue)
		}
	}

	return nil
}

func (r *RMQ) assertChannels() {
	if r.channels == nil {
		r.channels = make(map[string]*Channel)
	}
}

const (
	// declare flags
	Durable   = uint(0x1)
	Exclusive = uint(0x2)

	// publish flags
	Mandatory = uint(0x1)
	Immediate = uint(0x2)
)

func newChannel(r *RMQ, prefetchCount int) (*Channel, error) {
	c, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}

	c.Qos(prefetchCount, 0, false)
	return c, nil
}

func BasicConsume(c *Channel, queue string, ctag string, callback ConsumerCallback) error {
	deliveries, err := c.Consume(
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
			callback(newMessage(&d))
		}
	}()

	return nil
}

func BasicPublish(c *Channel, exchange string, rk string, body []byte, headers map[string]interface{}, flags uint) error {
	for k, v := range headers {
		switch fv := v.(type) {
		case int:
			headers[k] = int32(fv)
		}
	}
	err := c.Publish(
		exchange,
		rk,
		(flags & Mandatory) == Mandatory,
		(flags & Immediate) == Immediate,
		amqp.Publishing{
			Headers:      headers,
			Body:         body,
			DeliveryMode: amqp.Transient,
		},
	)
	if err != nil {
		return fmt.Errorf("Basic publish: %s", err)
	}

	return nil
}

func QueueBind(c *Channel, name, key, exchange string, args map[string]interface{}) error {
	return c.QueueBind(name, key, exchange, false, amqp.Table(args))
}

func ExchangeDeclare(c *Channel, exchange, typ string, flags uint) error {
	return c.ExchangeDeclare(
		exchange,
		typ,
		(flags & Durable) == Durable,
		false,
		false,
		false,
		nil,
	)
}

func QueueDeclare(c *Channel, queue string, flags uint) error {
	_, err := c.QueueDeclare(
		queue,
		(flags & Durable) == Durable,
		false,
		(flags & Exclusive) == Exclusive,
		false,
		nil,
	)
	return err
}

func Connect(host string) (*RMQ, error) {
	conn, err := amqp.Dial(host)
	if err != nil {
		return nil, err
	}

	ch := make(chan *amqp.Error)
	conn.NotifyClose(ch)

	go func() {
		for err := range ch {
			log.Println(err)
		}
	}()

	return &RMQ{
		channels: make(map[string]*Channel),
		connection: conn,
	}, nil
}

func ConnectAndAnnounce(sett *Settings) (*RMQ, error) {
	var r RMQ

	connection, err := amqp.Dial(sett.Amqp)
	if err != nil {
		return nil, err
	}

	ch := make(chan *amqp.Error)
	connection.NotifyClose(ch)

	go func() {
		for err := range ch {
			log.Println(err)
		}
	}()

	r.connection = connection
	channelAnnounce, err := r.Channel("announce", 0)
	if err != nil {
		return nil, err
	}

	err = r.announce(channelAnnounce, sett)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

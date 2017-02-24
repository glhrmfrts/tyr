package rmq

import (
	"fmt"
	"log"
	"github.com/streadway/amqp"
	"gitlab.com/vikingmakt/tyr/settings"
)

type RMQ struct {
	channels map[string]*Channel
	connection *amqp.Connection
}

func Connect(settings *settings.Settings) (*RMQ, error) {
	var r RMQ

	connection, err := amqp.Dial(settings.Amqp)
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

	err = r.announce(channelAnnounce, settings)
	if err != nil {
		return nil, err
	}

	return &r, nil
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
		channel.channel.NotifyClose(ch)

		go func() {
			for err := range ch {
				log.Println(err)
			}
		}()

		r.channels[name] = channel
	}

	return r.channels[name], nil
}

func (r *RMQ) announce(channel *Channel, settings *settings.Settings) error {
	var err error

	err = channel.channel.ExchangeDeclare(
		settings.Exchange.Topic,
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

	err = channel.channel.ExchangeDeclare(
		settings.Exchange.Headers,
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

	for _, s := range settings.Services {
		_, err = channel.channel.QueueDeclare(
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

		err = channel.channel.QueueBind(
			s.Queue,
			s.RoutingKey,
			settings.Exchange.Topic,
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

package rmq

import (
	"fmt"
	"log"
	"github.com/streadway/amqp"
	"gitlab.com/vikingmakt/tyr/ioengine"
	"gitlab.com/vikingmakt/tyr/settings"
)

type RMQ struct {
	channels map[string]*Channel
	connection *amqp.Connection
}

func (r *RMQ) Announce(settings *settings.Settings, future *ioengine.Future) {
	futureAnnounceChannel := ioengine.NewFuture(func(res ioengine.Result) {
		r.announce(res.Result.(*Channel), settings, future)
	})

	r.Channel("announce", 0, futureAnnounceChannel)
}

func (r *RMQ) Channel(name string, prefetchCount int, future *ioengine.Future) {
	go func() bool {
		r.assertChannels()

		if r.connection == nil {
			future.SetError(fmt.Errorf("No connection"))
		}

		if _, ok := r.channels[name]; !ok {
			channel, err := newChannel(r, prefetchCount)
			if err != nil {
				return future.SetError(err)
			}

			r.channels[name] = channel
		}

		return future.SetResult(r.channels[name])
	}()
}

func (r *RMQ) Connect(url string, future *ioengine.Future) {
	go func() bool {
		connection, err := amqp.Dial(url)
		if err != nil {
			return future.SetError(err)
		}

		r.connection = connection
		return future.SetResult(connection)
	}()
}

func (r *RMQ) announce(channel *Channel, settings *settings.Settings, future *ioengine.Future) {
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
		future.SetError(fmt.Errorf("Exchange Declare: %s", err))
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
		future.SetError(fmt.Errorf("Exchange Declare: %s", err))
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
			future.SetError(fmt.Errorf("Queue Declare: %s", err))
		}

		err = channel.channel.QueueBind(
			s.Queue,
			s.RoutingKey,
			settings.Exchange.Topic,
			false,      // noWait
			nil,        // arguments
		)
		if err != nil {
			future.SetError(fmt.Errorf("Queue Bind: %s", err))
		} else {
			log.Printf("RK %s -> Queue %s", s.RoutingKey, s.Queue)
		}
	}

	future.SetResult(true)
}

func (r *RMQ) assertChannels() {
	if r.channels == nil {
		r.channels = make(map[string]*Channel)
	}
}

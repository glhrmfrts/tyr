package rmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"gitlab.com/vikingmakt/tyr/ioengine"
)

type RMQ struct {
	channels map[string]*Channel
	connection *amqp.Connection
}

func (r *RMQ) Announce(settings *tyr.Settings, future *ioengine.Future) {
	futureAnnounceChannel := ioengine.NewFuture(func(res ioengine.Result) {
		r.announce(res.Result.(*Channel), settings, future)
	})

	r.Channel("announce", futureAnnounceChannel)
}

func (r *RMQ) Channel(name string, future *ioengine.Future) {
	go func() bool {
		r.assertChannels()

		if r.connection == nil {
			future.SetError(fmt.Errorf("No connection"))
		}

		channel, err := r.connection.Channel()
		if err != nil {
			return future.SetError(err)
		}

		r.channels[name] = &Channel{channel}
		return future.SetResult(channel)
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

func (r *RMQ) announce(channel *Channel, settings *tyr.Settings, future *ioengine.Future) {
	var err error

	err = channel.ExchangeDeclare(
		settings.Exchange.Topic,
		"topic",
		true,
		nil,
	)
	if err != nil {
		future.SetError(fmt.Errorf("Exchange Declare: %s", err))
	}

	err = channel.ExchangeDeclare(
		settings.Exchange.Headers,
		"headers",
		true,
		nil,
	)
	if err != nil {
		future.SetError(fmt.Errorf("Exchange Declare: %s", err))
	}

	for _, s := range settings.Services {
		err = channel.QueueDeclare(
			s.Queue,
			true,
			false,
			nil,
		)
		if err != nil {
			future.SetError(fmt.Errorf("Queue Declare: %s", err))
		}

		err = channel.QueueBind(
			settings.Exchange.Topic,
			s.Queue,
			s.RoutingKey,
		)
		if err != nil {
			future.SetError(fmt.Errorf("Queue Bind: %s", err))
		}
	}
}

func (r *RMQ) assertChannels() {
	if r.channels == nil {
		r.channels = make(map[string]*amqp.Channel)
	}
}

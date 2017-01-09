package rmq

import (
	"github.com/streadway/amqp"
)

type RMQ struct {
	channels map[string]*amqp.Channel
	connection *amqp.Connection
}
/*
func (r *RMQ) Channel(name string, future *ioengine.Future) {
	go func() {
		r.assertChannels()

		channel, err := r.connection.Channel()
		if err != nil {
			return future.SetError(err)
		}

		r.channels[name] = channel
		return future.SetResult(channel)
	}()
}

func (r *RMQ) Connect(url string, future *ioengine.Future) error {
	go func() {
		r.connection, err = amqp.Dial(url)
		if err != nil {
			return future.SetError(err)
		}

		return future.SetResult(true)
	}()
}

func (r *RMQ) assertChannels() {
	if r.channels == nil {
		r.channels = make(map[string]*amqp.Channel)
	}
}
*/

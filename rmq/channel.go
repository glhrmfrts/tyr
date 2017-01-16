package rmq

import (
	"github.com/streadway/amqp"
)

type Channel struct {
	ch *amqp.Channel
}

type ConsumerCallback func(channel *Channel, msg *Message)

func (self *Channel) BasicConsume(queue string, consumer ConsumerCallback) {

}

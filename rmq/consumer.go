package rmq

import (
	"github.com/satori/go.uuid"
)

type Consumer struct {
	tag string
}

type ConsumerCallback func(msg *Message)

func (c *Consumer) Tag() string {
	if c.tag == "" {
		c.tag = uuid.NewV4().String()
	}
	return c.tag
}

package rmq

import (
	"gitlab.com/vikingmakt/tyr/raid"
)

type Consumer struct {
	tag string
}

type ConsumerCallback func(msg *Message)

func (c *Consumer) Tag() string {
	if c.tag == "" {
		c.tag = raid.Etag()
	}
	return c.tag
}

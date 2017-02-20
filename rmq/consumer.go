package rmq

import (
)

type Consumer struct {
	tag string
}

func (c *Consumer) Tag() string {
	// TODO: generate tag from uuid

	if c.tag == "" {
		c.tag = "simple-consumer"
	}
	return c.tag
}

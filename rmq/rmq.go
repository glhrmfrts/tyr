package rmq

import (
	"gitlab.com/vikingmakt/tyr/rmq/connection"
)

type RMQ struct {
	
}

func (rmq *RMQ) Connect(url string) {
	println("RMQ")
	connection.Connection(&connection.Config{URL: "amqp://localhost"})
}

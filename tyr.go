package tyr

import (
	"gitlab.com/vikingmakt/tyr/ioengine"
	"gitlab.com/vikingmakt/tyr/rmq"
)

type Main interface {
	After()
	Before()
}

type Tyr struct {
	IOEngine ioengine.IOEngine
	RMQ rmq.RMQ
}

func (self *Tyr) Run(m Main) {
	m.Before()
	self.IOEngine.AddCallback(m.After)
	self.IOEngine.Start()
}

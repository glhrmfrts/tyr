package tyr

import (
	"gitlab.com/vikingmakt/tyr/ioengine"
	"gitlab.com/vikingmakt/tyr/rmq"
)

type Main interface {
	After(v interface{})
	Before()
}

type Tyr struct {
	IOEngine ioengine.IOEngine
	RMQ rmq.RMQ
}

func (self *Tyr) Run(m Main) {
	m.Before()
	self.IOEngine.AddCallback(m.After, nil)
	self.IOEngine.Start()
}

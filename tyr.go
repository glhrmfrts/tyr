package tyr

import (
	"gitlab.com/vikingmakt/tyr/ioengine"
	"gitlab.com/vikingmakt/tyr/rmq"
)

type Tyr struct {
	IOEngine *ioengine.IOEngine
	RMQ *rmq.RMQ
}


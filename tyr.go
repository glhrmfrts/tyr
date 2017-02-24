package tyr

import (
	"gitlab.com/vikingmakt/tyr/rmq"
  "time"
)

type Main interface {
	Start()
}

func Run(m Main) {
	m.Start()

  for {
    time.Sleep(100 * time.Millisecond)
  }
}

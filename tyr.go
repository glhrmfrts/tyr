package tyr

import (
	"time"
)

func Run(start func()) {
	start()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

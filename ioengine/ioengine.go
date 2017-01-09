package ioengine

import (
	"sync"
)

type event struct {
	callback func(interface{})
	payload interface{}
}

type IOEngine struct {
	WaitGroup sync.WaitGroup
	on bool
	pending []event
	queue chan event
}

func (self *IOEngine) AddCallback(callback func(interface{}), payload interface{}) {
	self.WaitGroup.Add(1)

	event := event{
		callback: callback,
		payload: payload,
	}

	if self.on {
		self.queue <- event
	} else {
		self.pending = append(self.pending, event)
	}
}

func (self *IOEngine) Start() {
	self.queue = make(chan event)
	self.on = true
	go self.loop()

	for _, event := range self.pending {
		self.queue <- event
	}
}

func (self *IOEngine) Status() bool {
	return self.on
}

func (self *IOEngine) Stop() {
	self.on = false
}

func (self *IOEngine) loop() {
	for self.on {
		event := <-self.queue
		go func() {
			event.callback(event.payload)
			self.WaitGroup.Done()
		}()
	}
}

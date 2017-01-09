package ioengine

import (
	"time"
)

type event struct {
	callback func(interface{})
	payload interface{}
}

type IOEngine struct {
	on bool
	pending []event
	queue chan event
}

func (self *IOEngine) AddCallback(callback func(interface{}), payload interface{}) {
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

func (self *IOEngine) Future(callback func(Result)) *Future {
	return NewFuture(callback)
}

func (self *IOEngine) Start() {
	self.queue = make(chan event)
	self.on = true
	go self.loop()

	for len(self.pending) > 0 {
		self.queue <- self.pending[0]
		self.pending = self.pending[1:]
	}

	for self.on {
		time.Sleep(100 * time.Millisecond)
	}
}

func (self *IOEngine) Status() bool {
	return self.on
}

func (self *IOEngine) Stop() {
	self.on = false
}

func (self *IOEngine) loop() {
	if self.on {
		event := <-self.queue
		go event.callback(event.payload)
		go self.loop()
	}
}

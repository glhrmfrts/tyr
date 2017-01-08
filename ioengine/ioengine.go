package ioengine

type IOEngine struct {
	on bool
}

func (self *IOEngine) Stop() {
	self.on = false
}

func (self *IOEngine) Start() {
	self.on = true
}

func (self IOEngine) Status() bool {
	return self.on
}

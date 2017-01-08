package ioengine

type IOEngine struct {
	on bool
}

func (self IOEngine) loop() {
	if self.on 
}

func (self *IOEngine) Stop() {
	self.on = false
}

func (self *IOEngine) Start() {
	self.on = true
	go self.loop()
}

func (self IOEngine) Status() bool {
	return self.on
}

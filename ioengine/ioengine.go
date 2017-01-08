package ioengine

type event struct {
	callback func
	payload interface{}
}

type IOEngine struct {
	on bool
}

func (self IOEngine) loop() bool {
	if self.on {
		event <- self.queue
		go event.callback(event.payload)		
		go self.IOEngine.loop()

		return true
	}
	
	return false
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

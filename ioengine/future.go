package ioengine

type Future struct {
	channel chan Result
}

type Result struct {
	Result interface{}
	Error error
}

func NewFuture(callback func(Result)) *Future {
	f := &Future{
		channel: make(chan Result),
	}

	go func() {
		callback(<-f.channel)
	}()

	return f
}

func (f *Future) Set(value interface{}, error error) bool {
	f.channel <- Result{
		Result: value,
		Error: error,
	}
	return true
}

func (f *Future) SetError(error error) bool {
	return f.Set(nil, error)
}

func (f *Future) SetResult(value interface{}) bool {
	return f.Set(value, nil)
}

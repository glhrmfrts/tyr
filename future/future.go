package future

type Default func(res interface{}, err error)

type Future struct {
	channel chan result
}

type result struct {
	result interface{}
	error  error
}

func F(callback Default) *Future {
	f := &Future{
		channel: make(chan result),
	}

	go func() {
		res := <-f.channel
		callback(res.result, res.error)
	}()

	return f
}

func (f *Future) Set(value interface{}, error error) bool {
	f.channel <- result{
		result: value,
		error: error,
	}
	return true
}

func (f *Future) SetError(error error) bool {
	return f.Set(nil, error)
}

func (f *Future) SetResult(value interface{}) bool {
	return f.Set(value, nil)
}

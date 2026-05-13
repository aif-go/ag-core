package future

import "fmt"

func NewFutureFunc(f func() (interface{}, error)) func() (interface{}, error) {
	var res interface{}
	var err error

	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		res, err = f()
	}()
	return func() (interface{}, error) {
		<-c
		return res, err
	}
}

func FutureCall(f func() (interface{}, error), callback func(interface{}, error)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				callback(nil, fmt.Errorf("panic: %v", r))
			}
		}()
		response, err := f()
		callback(response, err)
	}()
}

package future

import (
	"fmt"

	"github.com/panjf2000/ants/v2"
)

func NewFutureFunc(f func() (interface{}, error)) func() (interface{}, error) {
	var res interface{}
	var err error

	c := make(chan struct{}, 1)
	// go func() {
	submitErr := ants.Submit(func() {
		defer close(c)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		res, err = f()
	})
	if submitErr != nil {
		err = submitErr
		close(c)
	}
	return func() (interface{}, error) {
		<-c
		return res, err
	}
}

func FutureCall(f func() (interface{}, error), callback func(interface{}, error)) {
	err := ants.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				callback(nil, fmt.Errorf("panic: %v", r))
			}
		}()
		response, err := f()
		callback(response, err)
	})
	if err != nil {
		callback(nil, err)
	}
}

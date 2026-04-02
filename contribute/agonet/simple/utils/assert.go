package utils

import "fmt"

// Assert if nil != err
func Assert(err error, msg ...interface{}) {
	if nil != err {
		if len(msg) <= 0 {
			panic(err)
		} else {
			panic(fmt.Errorf("%w: %s", err, fmt.Sprint(msg...)))
		}
	}
}

// AssertIf exp
func AssertIf(exp bool, msg string, args ...interface{}) {
	if exp {
		panic(fmt.Errorf(msg, args...))
	}
}

// AssertLength check error
func AssertLength(n int, err error) int {
	if nil != err {
		panic(err)
	}
	return n
}

// AssertLong check error
func AssertLong(n int64, err error) int64 {
	if nil != err {
		panic(err)
	}
	return n
}

// AssertBytes check error
func AssertBytes(b []byte, err error) []byte {
	if nil != err {
		panic(err)
	}
	return b
}

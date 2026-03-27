package simple

import "fmt"

func AsException(ex interface{}) error {
	if nil != ex {
		if e, ok := ex.(error); ok {
			return e
		}
		return fmt.Errorf("%v", ex)
	}
	return nil
}

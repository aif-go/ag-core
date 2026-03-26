package simple

import "fmt"

func AsException(ex interface{}) Exception {
	if nil != ex {
		if e, ok := ex.(Exception); ok {
			return e
		}
		return fmt.Errorf("%v", ex)
	}
	return nil
}

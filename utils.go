package xrpc

import (
	"reflect"
)

func isErrorType(t reflect.Type) bool {
	return t.Implements(reflect.TypeOf((*error)(nil)).Elem())
}

// suitableMethod checks if the method is suitable for registration
// suitable method should have no return value or two return values
// the second return value should be of type error
// e.g.
//
//	func (s *Service) Method(args)
//	func (s *Service) Method(args) (reply, error)
func suitableMethod(mtype reflect.Type) bool {
	if mtype.NumOut() == 0 {
		return true
	}

	if mtype.NumOut() != 2 || !isErrorType(mtype.Out(1)) {
		return false
	}

	return true
}

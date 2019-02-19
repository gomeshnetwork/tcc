package xerrors

import (
	"errors"
	"fmt"
	"reflect"
)

// Error .
type Error interface {
	error              // mixin standard error interface
	CallStack() string // get call stack
	Cause() error      // error chain
}

// PrintStack print stack flag
var PrintStack = true

type facadeImpl struct {
}

func (facade *facadeImpl) Is(err, target error) bool {

	current := err

	for {

		if current == target {

			return true
		}

		e, ok := current.(Error)

		if !ok {
			return false
		}

		current = e.Cause()

		if current == nil {
			return false
		}
	}

}

func (facade *facadeImpl) As(err error, target interface{}) bool {

	errPointT := reflect.TypeOf(target)

	if errPointT.Kind() != reflect.Ptr {
		panic("target must be a point")
	}

	errT := errPointT.Elem()

	if errT.Kind() != reflect.Ptr && errT.Kind() != reflect.Interface {
		panic(fmt.Sprintf("invalid type: %s", errT.Kind()))
	}

	current := err

	for {
		currentT := reflect.TypeOf(current)

		if currentT == errT || currentT.Implements(errT) {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(current))
			return true
		}

		e, ok := current.(Error)

		if !ok {
			return false
		}

		current = e.Cause()

		if current == nil {
			return false
		}
	}
}

func init() {
	RegisterFacade("xerrors", &facadeImpl{})
}

// Is check if the err is target err
func Is(err, target error) (ok bool) {
	loopFacade(func(facade ErrorFacade) bool {

		ok = facade.Is(err, target)

		if ok {
			return true
		}

		return false
	})

	return
}

// As check if the err is target err
func As(err error, target interface{}) (ok bool) {
	loopFacade(func(facade ErrorFacade) bool {

		ok = facade.As(err, target)

		if ok {
			return true
		}

		return false
	})

	return
}

// Errorf .
func Errorf(fmtstring string, args ...interface{}) error {
	return NewStackError(3, fmt.Errorf(fmtstring, args...), nil)
}

// New .
func New(message string) error {
	return NewStackError(3, errors.New(message), nil)
}

// Wrapf .
func Wrapf(err error, fmtstring string, args ...interface{}) error {
	return NewStackError(3, fmt.Errorf(fmtstring, args...), err)
}

// Wrap .
func Wrap(err error, message string) error {
	return NewStackError(3, errors.New(message), err)
}

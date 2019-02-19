package injector

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/dynamicgo/slf4go"

	"github.com/dynamicgo/xerrors"
)

// Errors
var (
	ErrNotFound = errors.New("resource not found")
	ErrExists   = errors.New("already exists")
	ErrType     = errors.New("unexpect param type")
)

// Injector injector engine
type Injector interface {
	Register(name string, service interface{})
	Get(name string, service interface{}) bool
	Find(services interface{}) // get services
	Bind(service interface{}) error
}

type injectorImpl struct {
	sync.RWMutex
	slf4go.Logger
	services map[string]interface{}
}

// New create new injector context
func New() Injector {
	return &injectorImpl{
		Logger:   slf4go.Get("injector"),
		services: make(map[string]interface{}),
	}
}

func (injector *injectorImpl) Register(name string, service interface{}) {
	injector.Lock()
	defer injector.Unlock()

	_, ok := injector.services[name]

	if ok {
		err := xerrors.Wrapf(ErrExists, "service %s already exists", name)

		panic(err)
	}

	injector.services[name] = service
}

func (injector *injectorImpl) Get(name string, service interface{}) bool {
	injector.RLock()
	defer injector.RUnlock()

	serviceT := reflect.TypeOf(service)

	if serviceT.Kind() != reflect.Ptr {
		err := xerrors.Wrapf(ErrType, "expect ptr of interface or struct,got %s", serviceT)
		panic(err)
	}

	serviceT = serviceT.Elem()

	if serviceT.Kind() == reflect.Ptr {
		serviceT = serviceT.Elem()
	}

	if serviceT.Kind() != reflect.Struct && serviceT.Kind() != reflect.Interface {
		err := xerrors.Wrapf(ErrType, "expect ptr of interface or struct,got %s", serviceT)
		panic(err)
	}

	storaged, ok := injector.services[name]

	if !ok {
		return false
	}

	if serviceT.Kind() == reflect.Struct {
		if reflect.TypeOf(storaged).Elem() != serviceT {
			return false
		}
	} else {
		if !reflect.TypeOf(storaged).Implements(serviceT) {
			return false
		}
	}

	reflect.ValueOf(service).Elem().Set(reflect.ValueOf(storaged))

	return true
}

func (injector *injectorImpl) Find(services interface{}) {
	injector.RLock()
	defer injector.RUnlock()

	serviceT := reflect.TypeOf(services)

	if serviceT.Kind() != reflect.Ptr {
		err := xerrors.Wrapf(ErrType, "expect ptr of slice ,got %s", serviceT)
		panic(err)
	}

	serviceT = serviceT.Elem()

	if serviceT.Kind() == reflect.Ptr {
		serviceT = serviceT.Elem()
	}

	if serviceT.Kind() != reflect.Slice {
		err := xerrors.Wrapf(ErrType, "expect ptr of interface or struct,got %s", serviceT)
		panic(err)
	}

	elemT := serviceT.Elem()

	if elemT.Kind() == reflect.Ptr {
		elemT = elemT.Elem()
	}

	var storaged []interface{}

	if elemT.Kind() == reflect.Struct {
		for name, service := range injector.services {
			if reflect.TypeOf(service).Elem() == elemT {
				storaged = append(storaged, service)
				injector.InfoF("find service %s with name %s", serviceT, name)
			}
		}
	} else if elemT.Kind() == reflect.Interface {
		for name, service := range injector.services {
			storageT := reflect.TypeOf(service)

			if storageT.Implements(elemT) {
				storaged = append(storaged, service)
				injector.InfoF("find service %s implement %s with name %s", storageT, serviceT, name)
			}
		}
	} else {
		err := xerrors.Wrapf(ErrType, "slice element expect ptr of interface or struct,got %s", serviceT)
		panic(err)
	}

	if len(storaged) > 0 {
		sliceValue := reflect.MakeSlice(reflect.SliceOf(serviceT.Elem()), len(storaged), len(storaged))

		deferred := false

		if serviceT.Elem() == elemT && elemT.Kind() != reflect.Interface {
			deferred = true
		}

		for i := 0; i < len(storaged); i++ {

			if deferred {
				sliceValue.Index(i).Set(reflect.ValueOf(storaged[i]).Elem())
			} else {
				sliceValue.Index(i).Set(reflect.ValueOf(storaged[i]))
			}

		}

		reflect.ValueOf(services).Elem().Set(sliceValue)
	}

}

func (injector *injectorImpl) Bind(service interface{}) error {
	serviceT := reflect.TypeOf(service)

	if serviceT.Kind() != reflect.Ptr || serviceT.Elem().Kind() != reflect.Struct {
		err := xerrors.Wrapf(ErrType, "expect ptr of  struct,got %s", serviceT)
		panic(err)
	}

	serviceT = serviceT.Elem()

	serviceValue := reflect.ValueOf(service).Elem()

	for i := 0; i < serviceT.NumField(); i++ {

		field := serviceT.Field(i)

		tagStr, ok := field.Tag.Lookup("inject")

		if !ok {
			continue
		}

		if strings.ToTitle(field.Name[:1]) != field.Name[:1] {
			panic(fmt.Sprintf("inject filed must be export: %s", field.Name))
		}

		tag, err := injector.parseTag(tagStr)

		if err != nil {
			return err
		}

		if err := injector.executeInjectWithTag(tag, serviceValue.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

type injectTag struct {
	Name string
}

func (injector *injectorImpl) parseTag(tag string) (*injectTag, error) {
	return &injectTag{
		Name: tag,
	}, nil
}

func (injector *injectorImpl) executeInjectWithTag(tag *injectTag, fieldValue reflect.Value) error {
	fieldType := fieldValue.Type()

	if (fieldType.Kind() != reflect.Ptr || fieldType.Elem().Kind() != reflect.Struct) && fieldType.Kind() != reflect.Interface {
		panic("invalid inject field type, expect ptr of struct or interface")
	}

	if ok := injector.Get(tag.Name, fieldValue.Addr().Interface()); !ok {
		return xerrors.Wrapf(ErrNotFound, "inject object %s with type %s not found", tag.Name, fieldType)
	}

	return nil
}

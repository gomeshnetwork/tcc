package injector

import "sync"

var globalInjector Injector
var globalOnce sync.Once

func globalInit() {
	globalInjector = New()
}

// Register call global injector with register function
func Register(key string, val interface{}) {
	globalOnce.Do(globalInit)

	globalInjector.Register(key, val)
}

// Get call global injector with get function
func Get(key string, val interface{}) bool {
	globalOnce.Do(globalInit)

	return globalInjector.Get(key, val)
}

// Find call global injector with Find function
func Find(val interface{}) {
	globalOnce.Do(globalInit)

	globalInjector.Find(val)
}

// Bind .
func Bind(val interface{}) error {
	globalOnce.Do(globalInit)

	return globalInjector.Bind(val)
}

package slf4go

import config "github.com/dynamicgo/go-config"

type nullBackend struct {
}

func newNullBackend() LoggerFactory {
	return &nullBackend{}
}

func (backend *nullBackend) GetLogger(name string) Logger {
	return &nullLogger{name: name}
}

type nullLogger struct {
	name string
}

func (logger *nullLogger) SourceCodeLevel(level int) {

}

func (logger *nullLogger) GetName() string {
	return logger.name
}

func (logger *nullLogger) Trace(args ...interface{}) {

}

func (logger *nullLogger) TraceF(format string, args ...interface{}) {

}

func (logger *nullLogger) Debug(args ...interface{}) {

}

func (logger *nullLogger) DebugF(format string, args ...interface{}) {

}

func (logger *nullLogger) Info(args ...interface{}) {

}

func (logger *nullLogger) InfoF(format string, args ...interface{}) {

}

func (logger *nullLogger) Warn(args ...interface{}) {

}

func (logger *nullLogger) WarnF(format string, args ...interface{}) {

}

func (logger *nullLogger) Error(args ...interface{}) {

}

func (logger *nullLogger) ErrorF(format string, args ...interface{}) {

}

func (logger *nullLogger) Fatal(args ...interface{}) {

}

func (logger *nullLogger) FatalF(format string, args ...interface{}) {

}

func init() {
	println("[slf4go] register null backend")
	RegisterBackend("null", func(config config.Config) (LoggerFactory, error) {
		return newNullBackend(), nil
	})
}

package slf4go

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	config "github.com/dynamicgo/go-config"
)

type colorConsole struct {
	messages chan func()
}

var once sync.Once
var colorC *colorConsole

func newColorConsole() LoggerFactory {

	once.Do(func() {
		colorC = &colorConsole{
			messages: make(chan func(), 1000),
		}

		go colorC.runLoop()
	})

	return colorC
}

var mutex sync.Mutex

func (console *colorConsole) runLoop() {
	for f := range console.messages {
		mutex.Lock()
		f()
		mutex.Unlock()
	}
}

func (console *colorConsole) GetLogger(name string) Logger {
	return &colorConsoleLogger{name: name, messages: console.messages, codelevel: 3}
}

type colorConsoleLogger struct {
	name      string
	messages  chan func()
	codelevel int
}

func (logger *colorConsoleLogger) SourceCodeLevel(level int) {
	logger.codelevel = level
}

func (logger *colorConsoleLogger) GetName() string {
	return logger.name
}

func (logger *colorConsoleLogger) source() string {
	_, filename, line, _ := runtime.Caller(logger.codelevel)

	return fmt.Sprintf("%s:%d", filepath.Base(filename), line)
}

func (logger *colorConsoleLogger) Trace(args ...interface{}) {

	s := logger.source()

	logger.messages <- func() {
		tracef("[%s][%s][%s] TRACE ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		tracep(args...)
		tracep("\n")
	}

}

func (logger *colorConsoleLogger) TraceF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		tracef("[%s][%s][%s] TRACE ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		tracef(format, args...)
		tracep("\n")
	}
}

func (logger *colorConsoleLogger) Debug(args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		debugf("[%s][%s][%s] DEBUG ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		debugp(args...)
		debugp("\n")
	}
}

func (logger *colorConsoleLogger) DebugF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		debugf("[%s][%s][%s] DEBUG ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		debugf(format, args...)
		debugp("\n")
	}
}

func (logger *colorConsoleLogger) Info(args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		infof("[%s][%s][%s] INFO  ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		infop(args...)
		infop("\n")
	}
}

func (logger *colorConsoleLogger) InfoF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		infof("[%s][%s][%s] INFO  ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		infof(format, args...)
		infop("\n")
	}
}

func (logger *colorConsoleLogger) Warn(args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		warnf("[%s][%s][%s] WARN  ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		warnp(args...)
		warnp("\n")
	}
}

func (logger *colorConsoleLogger) WarnF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		warnf("[%s][%s][%s] WARN  ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		warnf(format, args...)
		warnp("\n")
	}
}

func (logger *colorConsoleLogger) Error(args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		errorf("[%s][%s][%s] ERROR ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		errorp(args...)
		errorp("\n")
	}
}

func (logger *colorConsoleLogger) ErrorF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		errorf("[%s][%s][%s] ERROR ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		errorf(format, args...)
		errorp("\n")
	}
}

func (logger *colorConsoleLogger) Fatal(args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		fatalf("[%s][%s][%s] FATAL ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		fatalp(args...)
		fatalp("\n")
	}
}

func (logger *colorConsoleLogger) FatalF(format string, args ...interface{}) {
	s := logger.source()
	logger.messages <- func() {
		fatalf("[%s][%s][%s] FATAL ", time.Now().Format("2006-01-02 15:04:05"), logger.name, s)
		fatalf(format, args...)
		fatalp("\n")
	}
}

func init() {
	println("[slf4go] register console backend")
	RegisterBackend("console", func(config config.Config) (LoggerFactory, error) {
		return newColorConsole(), nil
	})
}

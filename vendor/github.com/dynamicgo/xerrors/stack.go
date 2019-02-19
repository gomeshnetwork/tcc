package xerrors

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type stackError struct {
	err     error     // raw error
	prev    error     // parent error
	stackPC []uintptr // error stack
}

// NewStackError create new error with skip
func NewStackError(skip int, err, prev error) Error {

	pcs := make([]uintptr, 32)

	count := runtime.Callers(skip, pcs)

	return &stackError{
		err:     err,
		prev:    prev,
		stackPC: pcs[:count],
	}
}

func (err *stackError) Error() string {

	msg := err.err.Error()

	if PrintStack {
		msg = fmt.Sprintf("%s\n%s", msg, err.CallStack())
	}

	if err.prev != nil {
		msg = fmt.Sprintf("%scaused by: %s", msg, err.prev)
	}

	return msg
}

func (err *stackError) CallStack() string {
	frames := runtime.CallersFrames(err.stackPC)

	var buff bytes.Buffer

	for {
		frame, more := frames.Next()

		if index := strings.Index(frame.File, "src"); index != -1 {
			// trim GOPATH or GOROOT prifix
			frame.File = string(frame.File[index+4:])
		}

		buff.WriteString(fmt.Sprintf("\t%s(%s:%d)\n", frame.Function, filepath.Base(frame.File), frame.Line))

		if !more {
			break
		}
	}

	return buff.String()
}

func (err *stackError) Cause() error {
	return err.prev
}

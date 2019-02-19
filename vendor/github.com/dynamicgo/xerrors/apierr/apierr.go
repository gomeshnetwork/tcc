package apierr

import (
	"strings"

	"github.com/dynamicgo/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIErr .
type APIErr interface {
	error
	Code() int
	Scope() string
}

type apiErr struct {
	message string
	code    int
	scope   string
}

// New .
func New(code int, message string) APIErr {
	return &apiErr{
		message: message,
		code:    code,
		scope:   "default",
	}
}

// WithScope .
func WithScope(code int, message string, scope string) APIErr {
	return &apiErr{
		message: message,
		code:    code,
		scope:   scope,
	}
}

func (err *apiErr) Error() string {
	return err.message
}

func (err *apiErr) Code() int {
	return err.code
}

func (err *apiErr) Scope() string {
	return err.scope
}

// As convert any err to APIErr
func As(err error, deferr APIErr) APIErr {

	if err == nil {
		panic("invalid input")
	}

	var ae APIErr

	if xerrors.As(err, &ae) {
		return ae
	}

	s, ok := status.FromError(err)

	if ok && s.Code() > 5000 {

		messages := strings.Split(s.Message(), "|||")

		if len(messages) == 2 {
			return WithScope(-int(s.Code()-10000), messages[1], messages[0])
		}

		return New(-int(s.Code()-10000), s.Message())

	}

	return deferr
}

// AsGrpcError .
func AsGrpcError(err APIErr) error {

	code := uint32(-err.Code())

	return status.New(codes.Code(10000+code), err.Scope()+"|||"+err.Error()).Err()
}

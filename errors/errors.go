package errors

import "github.com/pkg/errors"

var (
	ErrGeneric          = errors.New("something went wrong")
	ErrCreateUserFailed = errors.New("failed to create user")
)

func New(message string) error {
	return errors.New(message)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

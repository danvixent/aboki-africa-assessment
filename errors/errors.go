package errors

import "github.com/pkg/errors"

var (
	ErrGeneric           = errors.New("something went wrong")
	ErrDebitUserFailed   = errors.New("failed to debit user")
	ErrCreditUserFailed  = errors.New("failed to credit user")
	ErrInsufficientFunds = errors.New("insufficient funds for the operation you're trying to perform")
	ErrCreateUserFailed  = errors.New("failed to create user")
)

func New(message string) error {
	return errors.New(message)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

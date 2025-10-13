package errutil

import (
	"fmt"

	l "github.com/sprisa/west/util/log"
)

// Panic (Fatal) error if condition is false
func Invariant(condition bool, format string, a ...any) {
	if !condition {
		l.Log.Panic().Msgf(format, a...)
	}
}

func InvariantError(err error, format string, a ...any) {
	if err != nil {
		l.Log.Panic().Err(err).Msgf(format, a...)
	}
}

func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s: %w", msg, err)
}

func Must[T any](v T, err error) *MustExec[T] {
	return &MustExec[T]{v, err}
}

type MustExec[T any] struct {
	v   T
	err error
}

func (s MustExec[T]) Run(format string, a ...any) T {
	if s.err != nil {
		l.Log.Panic().Err(s.err).Msgf(format, a...)
	}
	return s.v
}

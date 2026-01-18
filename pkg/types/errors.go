package types

import (
	"errors"
	"fmt"
	"time"
)

/*
sentinel base errors.
these are used with wrapping so callers can inspect errors via errors.Is.
*/
var (
	ErrTopicNotFound    = errors.New("topic not found")
	ErrInvalidMessage   = errors.New("invalid message")
	ErrConnectionClosed = errors.New("connection closed")
	ErrTimeout          = errors.New("timeout")
	ErrUnauthorized     = errors.New("unauthorized")
)

/*
wrapped error constructors.
use these to add context while preserving the base error.
*/

func NewTopicNotFoundError(topic string) error {
	return fmt.Errorf("topic %q: %w", topic, ErrTopicNotFound)
}
func NewInvalidMessageError(reason string) error {
	return fmt.Errorf("%s: %w", reason, ErrInvalidMessage)
}

func NewConnectionClosedError(op string) error {
	return fmt.Errorf("%s: %w", op, ErrConnectionClosed)
}

func NewTimeoutError(op string, d time.Duration) error {
	return fmt.Errorf("%s after %s: %w", op, d, ErrTimeout)
}

func NewUnauthorizedError(op string) error {
	return fmt.Errorf("%s: %w", op, ErrUnauthorized)
}

/*
helper functions for error classification.
these should be preferred over direct comparisons.
*/

func IsNotFound(err error) bool {
	return errors.Is(err, ErrTopicNotFound)
}

func IsInvalidMessage(err error) bool {
	return errors.Is(err, ErrInvalidMessage)
}

func IsConnectionClosed(err error) bool {
	return errors.Is(err, ErrConnectionClosed)
}

func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

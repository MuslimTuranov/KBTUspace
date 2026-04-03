package events

import "errors"

var (
	ErrEventFull           = errors.New("event is full")
	ErrAlreadyRegistered   = errors.New("already registered")
)


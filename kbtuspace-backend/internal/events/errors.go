package events

import "errors"

var (
	ErrEventFull         = errors.New("event is full")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrNotRegistered     = errors.New("not registered for this event")
	ErrForbidden         = errors.New("forbidden")
	ErrFacultyRequired   = errors.New("faculty is required")
	ErrApprovalPending   = errors.New("global content requires admin approval")
)

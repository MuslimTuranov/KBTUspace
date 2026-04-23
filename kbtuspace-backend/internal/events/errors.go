package events

import "errors"

var (
	ErrEventFull          = errors.New("event is full")
	ErrAlreadyRegistered  = errors.New("already registered")
	ErrNotRegistered      = errors.New("not registered for this event")
	ErrForbidden          = errors.New("forbidden")
	ErrCrossFacultyAccess = errors.New("cross-faculty access is forbidden")
	ErrInvalidEventDate   = errors.New("invalid event_date format")
	ErrFacultyRequired    = errors.New("faculty is required")
	ErrApprovalPending    = errors.New("global content requires admin approval")
)

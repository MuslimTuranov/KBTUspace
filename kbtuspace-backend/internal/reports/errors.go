package reports

import "errors"

var (
	ErrInvalidTargetType = errors.New("invalid target type")
	ErrTargetNotFound    = errors.New("target not found")
	ErrDuplicatePending  = errors.New("pending report already exists for this target")
	ErrSelfReport        = errors.New("you cannot report your own content")
)

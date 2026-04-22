package posts

import "errors"

var (
	ErrForbidden       = errors.New("forbidden")
	ErrPinForbidden    = errors.New("pinning posts is not allowed for this role")
	ErrFacultyRequired = errors.New("faculty is required")
	ErrApprovalPending = errors.New("global content requires admin approval")
)

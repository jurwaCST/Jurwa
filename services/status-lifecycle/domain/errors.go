package domain

import "errors"

var (
	// ErrInvalidTransition: the status flow does not allow this change.
	ErrInvalidTransition = errors.New("invalid status transition")
	// ErrInvalidStatus: the requested status is not one of the five statuses.
	ErrInvalidStatus = errors.New("invalid status")
	// ErrNotFound: no status is stored for this report.
	ErrNotFound = errors.New("report not found")
	// ErrAlreadyExists: the report already has a status.
	ErrAlreadyExists = errors.New("report already exists")
	// ErrVersionConflict: someone else changed the report first.
	ErrVersionConflict = errors.New("version conflict")
)

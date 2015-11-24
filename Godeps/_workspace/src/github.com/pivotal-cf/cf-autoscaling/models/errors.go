package models

import (
	"errors"
)

var (
	ErrRecordNotFound         = errors.New("Record Not Found")
	ErrDuplicateRecord        = errors.New("Duplicate Record")
	ErrNoScheduledEventsFound = errors.New("No future scheduled events found")
)

package event

import (
	"errors"
)

var WarnErrors = []error{
	ErrEventNotFound,
	ErrRequiredField,
}

var (
	// Event
	ErrEventNotFound               = errors.New("event not found")
	ErrRequiredField               = errors.New("required field")
	ErrTimezoneNotFound            = errors.New("timezone not found")
	ErrEventEditNotAllowed         = errors.New("only the user who created the event can edit it")
	ErrDepartmentNotBelongToBranch = errors.New("department not belong to branch")
	ErrAssignNotBelongToBranch     = errors.New("assign not belong to branch")
	ErrCanNotViewEvent             = errors.New("can not view event")
	ErrRoomUnavailable             = errors.New("room unavailable")
)

package http

import (
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	pkgErrors "gitlab.com/gma-vietnam/tanca-connect/pkg/errors"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// Event errors
	errEventNotFound               = pkgErrors.NewHTTPError(141004, "Event not found")
	errRequiredField               = pkgErrors.NewHTTPError(141005, "Required field")
	errTimezoneNotFound            = pkgErrors.NewHTTPError(141006, "Timezone not found")
	errAssignNotBelongToBranch     = pkgErrors.NewHTTPError(141007, "Assign not belong to branch")
	errDepartmentNotBelongToBranch = pkgErrors.NewHTTPError(141008, "Department not belong to branch")
	errCanNotViewEvent             = pkgErrors.NewHTTPError(141009, "Can not view event")

	// Shop service
	errElementNotFound     = pkgErrors.NewHTTPError(140004, "Element not found")
	errShopElementNotFound = pkgErrors.NewHTTPError(140005, "Shop element not found")
	errDepartmentNotFound  = pkgErrors.NewHTTPError(140006, "Department not found")
	errBranchNotFound      = pkgErrors.NewHTTPError(140007, "Branch not found")

	// Room service
	errRoomUnavailable = pkgErrors.NewHTTPError(140008, "Room unavailable")
)

func (h handler) mapError(err error) error {
	switch err {
	// Event errors
	case event.ErrEventNotFound:
		return errEventNotFound
	case event.ErrRequiredField:
		return errRequiredField
	case event.ErrTimezoneNotFound:
		return errTimezoneNotFound
	case event.ErrAssignNotBelongToBranch:
		return errAssignNotBelongToBranch
	case event.ErrDepartmentNotBelongToBranch:
		return errDepartmentNotBelongToBranch
	case event.ErrCanNotViewEvent:
		return errCanNotViewEvent

	// Shop service
	case microservice.ErrElementNotFound:
		return errElementNotFound
	case microservice.ErrShopElementNotFound:
		return errShopElementNotFound
	case microservice.ErrDepartmentNotFound:
		return errDepartmentNotFound
	case microservice.ErrBranchNotFound:
		return errBranchNotFound

	default:
		panic(err)
	}
}

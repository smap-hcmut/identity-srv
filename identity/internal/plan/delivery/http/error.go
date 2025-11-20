package http

import (
	"smap-api/internal/plan"
	pkgErrors "smap-api/pkg/errors"
)

var (
	errWrongBody        = pkgErrors.NewHTTPError(120001, "Wrong body")
	errPlanNotFound     = pkgErrors.NewHTTPError(120002, "Plan not found")
	errPlanExists       = pkgErrors.NewHTTPError(120003, "Plan already exists")
	errInvalidPlan      = pkgErrors.NewHTTPError(120004, "Invalid plan")
	errFieldRequired    = pkgErrors.NewHTTPError(120005, "Field required")
	errPlanCodeExists   = pkgErrors.NewHTTPError(120006, "Plan code already exists")
	errInvalidID        = pkgErrors.NewHTTPError(120007, "Invalid ID")
)

func (h handler) mapErrorCode(err error) error {
	switch err {
	case errWrongBody:
		return errWrongBody
	case plan.ErrPlanNotFound:
		return errPlanNotFound
	case plan.ErrPlanExists:
		return errPlanExists
	case plan.ErrInvalidPlan:
		return errInvalidPlan
	case plan.ErrFieldRequired:
		return errFieldRequired
	case plan.ErrPlanCodeExists:
		return errPlanCodeExists
	case errInvalidID:
		return errInvalidID
	default:
		return err
	}
}

var NotFound = []error{
	errPlanNotFound,
}


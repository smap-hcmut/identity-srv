package plan

import "errors"

var (
	ErrPlanNotFound   = errors.New("plan not found")
	ErrPlanExists     = errors.New("plan already exists")
	ErrInvalidPlan    = errors.New("invalid plan")
	ErrFieldRequired  = errors.New("field required")
	ErrPlanCodeExists = errors.New("plan code already exists")
)

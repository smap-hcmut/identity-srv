package audit

import "errors"

var (
	ErrPublishFailed = errors.New("failed to publish audit event")
)

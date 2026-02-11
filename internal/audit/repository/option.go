package repository

import "time"

// QueryOptions contains options for querying audit logs
type QueryOptions struct {
	UserID string
	Action string
	From   *time.Time
	To     *time.Time
	Page   int
	Limit  int
}

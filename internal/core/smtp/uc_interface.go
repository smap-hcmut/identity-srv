package smtp

import "context"

type UseCase interface {
	SendEmail(ctx context.Context, data EmailData) error
}

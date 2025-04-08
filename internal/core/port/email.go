package port

import (
	"context"
)

type EmailRepository interface {
	// Set stores the value in the cache
	SendEmail(ctx context.Context, from string, to string, subject string, content string) error
}

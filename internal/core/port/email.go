package port

import (
	"context"
)

type EmailRepository interface {
	SendEmail(ctx context.Context, to[] string, subjects string, textContent string, htmlContent string) error
}

package email_test


import (
	"context"
	"harajuku/backend/internal/adapter/communication/email"
	"harajuku/backend/internal/adapter/config"
	"testing"
)

const URL = ""
const API_TOKEN = ""
const FROM_EMAIL = ""

func TestSendEmailIntegrationTest(t *testing.T) {
    ctx := context.Background()

    cfg := &config.Email{
        Url:       URL,
        ApiToken:  API_TOKEN,
        FromEmail: FROM_EMAIL,
    }

    emailManager, err := email.New(ctx, cfg)
    if err != nil {
        t.Fatalf("failed to create the Email Manager: %v", err)
    }

    t.Logf("Using endpoint: %s", URL) // Debug output

    emails := []string{"gizehmata@gmail.com", "shely0210@hotmail.com"}

    err = emailManager.SendEmail(ctx, emails, "Saludo", "Hola, cómo estás?", "")
    if err != nil {
        t.Fatalf("failed to send email: %v", err) // Updated error message
    }
}

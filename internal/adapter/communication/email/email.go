package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/core/port"
	"net/http"
	"time"
)

type EmailManager struct {
	BaseURL   string
	APIToken  string
	FromEmail string
	Client    *http.Client
}

type mailtrapRecipient struct {
	Email string `json:"email"`
}

type mailtrapPayload struct {
	From struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"from"`
	To      []mailtrapRecipient `json:"to"`
	Subject string              `json:"subject"`
	Text    string              `json:"text,omitempty"`
	HTML    string              `json:"html,omitempty"`
}

func New(ctx context.Context, config *config.Email) (port.EmailRepository, error) {
	if config == nil || config.Url == "" || config.ApiToken == "" {
		return nil, errors.New("invalid email configuration")
	}

	return &EmailManager{
		BaseURL:   config.Url,
		APIToken:  config.ApiToken,
		FromEmail: config.FromEmail,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (em *EmailManager) SendEmail(ctx context.Context, to []string, subject string, textContent string, htmlContent string) error {
	// Validate at least one recipient
	if len(to) == 0 {
		return errors.New("at least one recipient is required")
	}

	// Validate at least one content type is provided
	if textContent == "" && htmlContent == "" {
		return errors.New("either text or HTML content must be provided")
	}

	// Prepare payload
	payload := mailtrapPayload{}
	payload.From.Email = em.FromEmail
	payload.From.Name = "Harajuku"
	payload.Subject = subject

	// Add recipients
	for _, recipient := range to {
		payload.To = append(payload.To, mailtrapRecipient{
			Email: recipient,
		})
	}

	// Set content based on what's provided
	if textContent != "" {
		payload.Text = textContent
	}
	if htmlContent != "" {
		payload.HTML = htmlContent
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		em.BaseURL,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", em.APIToken))
	req.Header.Add("Content-Type", "application/json")


	res, err := em.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer res.Body.Close()


	if res.StatusCode >= 400 {
		return fmt.Errorf("email sending failed with status: %d", res.StatusCode)
	}

	return nil
}

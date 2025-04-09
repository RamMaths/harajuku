package email

import (
	"context"
	"testing"

	"harajuku/backend/internal/adapter/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		cfg := &config.Email{
			Url:      "https://api.mailtrap.io",
			ApiToken: "test-token",
			FromEmail:     "test@example.com",
		}

		repo, err := New(context.Background(), cfg)
		require.NoError(t, err)
		assert.NotNil(t, repo)
		
		// Type assert to access manager fields
		manager, ok := repo.(*EmailManager)
		require.True(t, ok)
		assert.Equal(t, "https://api.mailtrap.io", manager.BaseURL)
		assert.Equal(t, "test-token", manager.APIToken)
	})

	t.Run("missing configuration", func(t *testing.T) {
		_, err := New(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("missing URL", func(t *testing.T) {
		cfg := &config.Email{
			ApiToken: "test-token",
			FromEmail:     "test@example.com",
		}
		_, err := New(context.Background(), cfg)
		assert.Error(t, err)
	})
}


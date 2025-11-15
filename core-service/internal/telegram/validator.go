package telegram

import (
	"context"
	"time"

	"core-service/internal/domain"
	"core-service/internal/logger"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// UserPayload describes trimmed Telegram user object.
type UserPayload struct {
	ID           int64  `json:"id"`
	LanguageCode string `json:"language_code"`
	PhotoURL     string `json:"photo_url"`
}

// InitData represents validated Telegram init data.
type InitData struct {
	User     UserPayload
	AuthDate time.Time
	QueryID  string
}

// Validator validates Telegram init data strings coming from TMA.
type Validator struct {
	botTokens []string
	maxAge    time.Duration
}

// NewValidator creates a new validator instance.
func NewValidator(botTokens []string, maxAge time.Duration) *Validator {
	return &Validator{botTokens: botTokens, maxAge: maxAge}
}

// Validate parses and validates init data. Returns ErrUnauthorized if invalid.
func (v *Validator) Validate(initDataStr string) (*InitData, error) {
	ctx := context.Background()

	if initDataStr == "" {
		logger.Warn(ctx, "telegram validator: empty init data")
		return nil, domain.ErrUnauthorized
	}

	// Validate using the library - try all tokens
	var lastErr error
	for _, token := range v.botTokens {
		if valErr := initdata.Validate(initDataStr, token, v.maxAge); valErr == nil {
			// Valid token found, continue with parsing
			goto parseinitdata
		} else {
			lastErr = valErr
		}
	}
	// None of the tokens validated successfully
	logger.Warnf(ctx, "telegram validator: validation failed with all tokens: %v", lastErr)
	return nil, domain.ErrUnauthorized

parseinitdata:

	// Parse init data
	data, err := initdata.Parse(initDataStr)
	if err != nil {
		logger.Warnf(ctx, "telegram validator: failed to parse init data: %v", err)
		return nil, domain.ErrUnauthorized
	}

	// Map to our structure
	result := &InitData{
		AuthDate: data.AuthDate(),
		QueryID:  data.QueryID,
	}

	// Extract user data
	if data.User.ID != 0 {
		result.User = UserPayload{
			ID:           data.User.ID,
			LanguageCode: data.User.LanguageCode,
			PhotoURL:     data.User.PhotoURL,
		}
	} else {
		logger.Warn(ctx, "telegram validator: user data not found in init data")
		return nil, domain.ErrUnauthorized
	}

	return result, nil
}

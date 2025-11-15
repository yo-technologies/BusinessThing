package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"core-service/internal/domain"
	"core-service/internal/logger"
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
	botToken string
	maxAge   time.Duration
}

// NewValidator creates a new validator instance.
func NewValidator(botToken string, maxAge time.Duration) *Validator {
	return &Validator{botToken: botToken, maxAge: maxAge}
}

// Validate parses and validates init data. Returns ErrUnauthorized if invalid.
func (v *Validator) Validate(initData string) (*InitData, error) {
	ctx := context.Background()
	
	if initData == "" {
		logger.Warn(ctx, "telegram validator: empty init data")
		return nil, domain.ErrUnauthorized
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		logger.Warnf(ctx, "telegram validator: failed to parse query: %v", err)
		return nil, domain.ErrUnauthorized
	}

	hash := values.Get("hash")
	if hash == "" {
		logger.Warn(ctx, "telegram validator: hash not found in init data")
		return nil, domain.ErrUnauthorized
	}

	values.Del("hash")

	dataCheckString := buildDataCheckString(values)
	if !v.verifyHash(dataCheckString, hash) {
		logger.Warn(ctx, "telegram validator: hash verification failed")
		return nil, domain.ErrUnauthorized
	}

	authDateUnix, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		logger.Warnf(ctx, "telegram validator: failed to parse auth_date: %v", err)
		return nil, domain.ErrUnauthorized
	}
	authDate := time.Unix(authDateUnix, 0)
	if v.maxAge > 0 && time.Since(authDate) > v.maxAge {
		logger.Warnf(ctx, "telegram validator: init data expired (age: %v, max: %v)", time.Since(authDate), v.maxAge)
		return nil, domain.ErrUnauthorized
	}

	var userPayload UserPayload
	if err := json.Unmarshal([]byte(values.Get("user")), &userPayload); err != nil {
		logger.Warnf(ctx, "telegram validator: failed to unmarshal user payload: %v", err)
		return nil, domain.ErrUnauthorized
	}

	return &InitData{
		User:     userPayload,
		AuthDate: authDate,
		QueryID:  values.Get("query_id"),
	}, nil
}

func buildDataCheckString(values url.Values) string {
	parts := make([]string, 0, len(values))
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, vals[0]))
	}
	sort.Strings(parts)
	return strings.Join(parts, "\n")
}

func (v *Validator) verifyHash(dataCheckString, hash string) bool {
	if v.botToken == "" {
		return false
	}

	secretKeyMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyMAC.Write([]byte(v.botToken))
	secretKey := secretKeyMAC.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	expected := mac.Sum(nil)

	actual, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	return hmac.Equal(expected, actual)
}

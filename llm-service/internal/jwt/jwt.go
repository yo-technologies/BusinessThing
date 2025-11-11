package jwt

import (
	"context"
	"crypto/rsa"
	"llm-service/internal/domain"
	"llm-service/internal/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type credentials interface {
	NewWithClaims(claims jwt.Claims) (string, error)
	GetKey(*jwt.Token) (interface{}, error)
}

type PemCredentials struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func NewPemCredentials(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *PemCredentials {
	return &PemCredentials{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

func (c *PemCredentials) NewWithClaims(claims jwt.Claims) (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(c.PrivateKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *PemCredentials) GetKey(token *jwt.Token) (interface{}, error) {
	return c.PublicKey, nil
}

type SecretCredentials struct {
	Secret string
}

func NewSecretCredentials(secret string) *SecretCredentials {
	return &SecretCredentials{
		Secret: secret,
	}
}

func (c *SecretCredentials) NewWithClaims(claims jwt.Claims) (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(c.Secret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *SecretCredentials) GetKey(token *jwt.Token) (interface{}, error) {
	return []byte(c.Secret), nil
}

type Options struct {
	Credentials credentials

	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type OptionsFunc func(*Options)

func WithCredentials(credentials credentials) OptionsFunc {
	return func(o *Options) {
		o.Credentials = credentials
	}
}

func WithAccessTTL(ttl time.Duration) OptionsFunc {
	return func(o *Options) {
		o.AccessTTL = ttl
	}
}

func WithRefreshTTL(ttl time.Duration) OptionsFunc {
	return func(o *Options) {
		o.RefreshTTL = ttl
	}
}

type Provider struct {
	opts Options
}

func NewProvider(opts ...OptionsFunc) *Provider {
	p := &Provider{
		opts: Options{
			AccessTTL:  time.Hour,
			RefreshTTL: time.Hour * 24 * 7,
		},
	}

	for _, o := range opts {
		o(&p.opts)
	}

	if p.opts.Credentials == nil {
		logger.Panic(context.Background(), "credentials are required")
	}

	return p
}

func (p *Provider) ParseToken(ctx context.Context, token string) (domain.ID, error) {
	claims := jwt.MapClaims{}
	tokenParsed, err := jwt.ParseWithClaims(token, &claims, p.opts.Credentials.GetKey)
	if err != nil {
		logger.Errorf(ctx, "failed to parse token: %v", err)
		return domain.ID{}, domain.ErrUnauthorized
	}

	if !tokenParsed.Valid {
		logger.Errorf(ctx, "token is invalid")
		return domain.ID{}, domain.ErrUnauthorized
	}

	subject, err := claims.GetSubject()
	if err != nil {
		logger.Errorf(ctx, "token subject not found: %v", err)
		return domain.ID{}, domain.ErrUnauthorized
	}

	return domain.ParseID(subject)
}

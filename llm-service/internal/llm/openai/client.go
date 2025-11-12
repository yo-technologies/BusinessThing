package openai_llm

import (
	"llm-service/internal/config"
	"llm-service/internal/llm"

	"github.com/openai/openai-go/v3"
)

// Ensure implementation
var _ llm.CompletionProvider = (*CompletionProvider)(nil)

type CompletionProvider struct {
	client openai.Client
	cfg    *config.Config
}

func New(client openai.Client, cfg *config.Config) *CompletionProvider {
	return &CompletionProvider{
		client: client,
		cfg:    cfg,
	}
}

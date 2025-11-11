package chat

import (
	"context"
	"strings"
	"time"

	"llm-service/internal/domain"

	"github.com/opentracing/opentracing-go"
)

// buildSystemPrompt returns the system prompt enriched with short user facts (if available).
func (s *Service) buildSystemPrompt(ctx context.Context, userID domain.ID) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.chat.buildSystemPrompt")
	defer span.Finish()

	system := strings.Builder{}
	system.WriteString(defaultChatSystemPrompt)

	facts, err := s.memoryService.ListFacts(ctx, userID)
	if err != nil {
		return system.String()
	}

	if len(facts) > 0 {
		system.WriteString("\n\nКраткие факты о пользователе (помни и учитывай в ответах):\n")
		for _, f := range facts {
			system.WriteString("- ")
			system.WriteString(f.Content)
			system.WriteString("\n")
		}

	}

	system.WriteString("\n\nТекущее время: ")
	system.WriteString(time.Now().Format(time.RFC1123))

	return system.String()
}

package memory

import (
	"context"
	"strings"

	"llm-service/internal/domain"

	"github.com/opentracing/opentracing-go"
)

const (
	// MaxFactsPerUser ограничение количества фактов на пользователя
	MaxFactsPerUser = 50
	// MaxFactLength максимальная длина одного факта (символов)
	MaxFactLength = 200
)

type repository interface {
	CreateUserMemoryFact(ctx context.Context, fact domain.UserMemoryFact) (domain.UserMemoryFact, error)
	ListUserMemoryFacts(ctx context.Context, userID domain.ID, limit int) ([]domain.UserMemoryFact, error)
	CountUserMemoryFacts(ctx context.Context, userID domain.ID) (int, error)
	DeleteUserMemoryFact(ctx context.Context, userID domain.ID, factID domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service { return &Service{repo: repo} }

// AddFact добавляет короткий факт для пользователя с валидацией и ограничениями.
func (s *Service) AddFact(ctx context.Context, userID domain.ID, content string) (domain.UserMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.memory.AddFact")
	defer span.Finish()

	content = strings.TrimSpace(content)
	if content == "" {
		return domain.UserMemoryFact{}, domain.ErrInvalidArgument
	}
	if len([]rune(content)) > MaxFactLength {
		content = string([]rune(content)[:MaxFactLength])
	}

	cnt, err := s.repo.CountUserMemoryFacts(ctx, userID)
	if err != nil {
		return domain.UserMemoryFact{}, err
	}
	if cnt >= MaxFactsPerUser {
		return domain.UserMemoryFact{}, domain.ErrTooManyRequests
	}

	fact := domain.NewUserMemoryFact(userID, content)

	created, err := s.repo.CreateUserMemoryFact(ctx, fact)
	if err != nil {
		return domain.UserMemoryFact{}, err
	}
	if created.ID == (domain.ID{}) {
		return created, nil
	}

	return created, nil
}

// ListFacts возвращает все факты пользователя с ограничением сверху.
func (s *Service) ListFacts(ctx context.Context, userID domain.ID) ([]domain.UserMemoryFact, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.memory.ListFacts")
	defer span.Finish()

	return s.repo.ListUserMemoryFacts(ctx, userID, MaxFactsPerUser)
}

// DeleteFact удаляет факт пользователя по ID.
func (s *Service) DeleteFact(ctx context.Context, userID domain.ID, factID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.memory.DeleteFact")
	defer span.Finish()

	return s.repo.DeleteUserMemoryFact(ctx, userID, factID)
}

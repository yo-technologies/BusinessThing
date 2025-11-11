package repository

import (
	"llm-service/internal/db"
)

type PGXRepository struct {
	engineFactory db.EngineFactory
}

func NewPGXRepository(engineFactory db.EngineFactory) *PGXRepository {
	return &PGXRepository{
		engineFactory: engineFactory,
	}
}

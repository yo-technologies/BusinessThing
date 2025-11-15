package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) IsEmpty() bool {
	return uuid.UUID(id) == uuid.Nil
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %w", err, ErrInvalidArgument)
	}
	return ID(id), nil
}

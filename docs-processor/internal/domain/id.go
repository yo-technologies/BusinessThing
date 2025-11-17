package domain

import (
	"encoding/json"
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

func (id ID) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(id).String()), nil
}

func (id *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*id = ID{}
		return nil
	}

	parsed, err := uuid.ParseBytes(text)
	if err != nil {
		return err
	}

	*id = ID(parsed)
	return nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	text, err := id.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

func (id *ID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = ID{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return id.UnmarshalText([]byte(s))
}

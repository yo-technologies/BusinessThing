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

func (i ID) String() string {
	return uuid.UUID(i).String()
}

func (i ID) MarshalText() ([]byte, error) {
	return []byte(uuid.UUID(i).String()), nil
}

func (i *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*i = ID{}
		return nil
	}

	parsed, err := uuid.ParseBytes(text)
	if err != nil {
		return err
	}

	*i = ID(parsed)
	return nil
}

func (i ID) MarshalJSON() ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

func (i *ID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*i = ID{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return i.UnmarshalText([]byte(s))
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %w", err, ErrInvalidArgument)
	}
	return ID(id), nil
}

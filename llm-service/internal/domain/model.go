package domain

import "time"

type Model struct {
	ID        ID        `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewModel() Model {
	return Model{
		ID:        NewID(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

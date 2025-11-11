package domain

// UserMemoryFact — короткий факт о пользователе, сохраняемый агентом.
// Используется для персонализации ответов и контекста.
type UserMemoryFact struct {
	Model
	UserID  ID     `db:"user_id"`
	Content string `db:"content"`
}

func NewUserMemoryFact(userID ID, content string) UserMemoryFact {
	return UserMemoryFact{
		Model:   NewModel(),
		UserID:  userID,
		Content: content,
	}
}

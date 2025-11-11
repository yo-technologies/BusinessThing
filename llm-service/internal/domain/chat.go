package domain

type Chat struct {
	Model
	UserID ID     `db:"user_id"`
	Title  string `db:"title"`
}

// NewChat constructs a chat with required fields
func NewChat(userID ID, title string) Chat {
	return Chat{
		Model:  NewModel(),
		UserID: userID,
		Title:  title,
	}
}

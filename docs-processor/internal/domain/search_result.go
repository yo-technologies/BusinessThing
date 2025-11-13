package domain

type SearchResult struct {
	ChunkID      ID
	DocumentID   ID
	DocumentName string
	Content      string
	Position     int
	Score        float32
	Metadata     map[string]string
}

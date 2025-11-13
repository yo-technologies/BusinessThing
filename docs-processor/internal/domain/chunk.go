package domain

type Chunk struct {
	ID         ID
	DocumentID ID
	Content    string
	Position   int
	Embedding  []float32
	Metadata   map[string]string
}

func NewChunk(documentID ID, content string, position int) *Chunk {
	return &Chunk{
		ID:         NewID(),
		DocumentID: documentID,
		Content:    content,
		Position:   position,
		Metadata:   make(map[string]string),
	}
}

func (c *Chunk) WithMetadata(key, value string) *Chunk {
	c.Metadata[key] = value
	return c
}

func (c *Chunk) WithEmbedding(embedding []float32) *Chunk {
	c.Embedding = embedding
	return c
}

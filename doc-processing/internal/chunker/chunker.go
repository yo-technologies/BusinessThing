package chunker

import (
	"context"
	"strings"

	"doc-processing/internal/domain"

	"github.com/opentracing/opentracing-go"
)

type Chunker struct {
	maxChunkSize int
	overlapSize  int
}

func New(maxChunkSize, overlapSize int) *Chunker {
	return &Chunker{
		maxChunkSize: maxChunkSize,
		overlapSize:  overlapSize,
	}
}

func (c *Chunker) ChunkText(ctx context.Context, documentID domain.ID, text string) ([]*domain.Chunk, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "chunker.Chunker.ChunkText")
	defer span.Finish()

	if text == "" {
		return []*domain.Chunk{}, nil
	}

	sentences := c.splitIntoSentences(text)
	chunks := make([]*domain.Chunk, 0)

	currentChunk := strings.Builder{}
	position := 0

	for i, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		if currentChunk.Len()+len(sentence) > c.maxChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, domain.NewChunk(documentID, currentChunk.String(), position))
			position++

			overlapText := c.getOverlapText(sentences, i, c.overlapSize)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, domain.NewChunk(documentID, currentChunk.String(), position))
	}

	return chunks, nil
}

func (c *Chunker) splitIntoSentences(text string) []string {
	sentences := make([]string, 0)
	currentSentence := strings.Builder{}

	for _, r := range text {
		currentSentence.WriteRune(r)

		if r == '.' || r == '!' || r == '?' || r == '\n' {
			sentence := strings.TrimSpace(currentSentence.String())
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			currentSentence.Reset()
		}
	}

	if currentSentence.Len() > 0 {
		sentence := strings.TrimSpace(currentSentence.String())
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}

func (c *Chunker) getOverlapText(sentences []string, startIdx, maxOverlapSize int) string {
	overlap := strings.Builder{}
	overlapChars := 0

	for i := startIdx - 1; i >= 0 && overlapChars < maxOverlapSize; i-- {
		sentence := sentences[i]
		if overlapChars+len(sentence) > maxOverlapSize {
			break
		}

		if overlap.Len() > 0 {
			overlap.WriteString(" ")
		}
		overlap.WriteString(sentence)
		overlapChars += len(sentence)
	}

	return reverseWords(overlap.String())
}

func reverseWords(s string) string {
	words := strings.Fields(s)
	for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
		words[i], words[j] = words[j], words[i]
	}
	return strings.Join(words, " ")
}

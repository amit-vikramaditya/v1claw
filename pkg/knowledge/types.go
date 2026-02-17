package knowledge

import (
	"context"
	"fmt"
	"math"
)

// EmbeddingProvider generates vector embeddings from text.
type EmbeddingProvider interface {
	// Embed generates a vector embedding for the given text.
	Embed(ctx context.Context, text string) ([]float64, error)
	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
	// Dimensions returns the embedding vector size.
	Dimensions() int
	// Name returns the provider name.
	Name() string
}

// Document represents a piece of knowledge stored in the system.
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Source    string            `json:"source"`    // e.g., "file:/path", "chat:session", "url:..."
	Metadata  map[string]string `json:"metadata,omitempty"`
	Embedding []float64         `json:"embedding,omitempty"`
}

// SearchResult is a document with similarity score.
type SearchResult struct {
	Document   Document `json:"document"`
	Score      float64  `json:"score"` // 0.0 to 1.0 (cosine similarity)
	Rank       int      `json:"rank"`
}

// ChunkOptions controls how documents are split.
type ChunkOptions struct {
	MaxChunkSize int // Max characters per chunk (default 512).
	Overlap      int // Characters overlap between chunks (default 64).
}

// DefaultChunkOptions returns sensible defaults.
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		MaxChunkSize: 512,
		Overlap:      64,
	}
}

// CosineSimilarity computes similarity between two vectors.
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimension mismatch: %d vs %d", len(a), len(b))
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0, nil
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

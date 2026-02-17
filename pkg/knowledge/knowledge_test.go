package knowledge

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEmbedder returns deterministic embeddings based on text content.
type mockEmbedder struct {
	dims int
}

func (m *mockEmbedder) Name() string       { return "mock" }
func (m *mockEmbedder) Dimensions() int     { return m.dims }
func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	results, err := m.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}
func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, text := range texts {
		vec := make([]float64, m.dims)
		for j := 0; j < m.dims && j < len(text); j++ {
			vec[j] = float64(text[j]) / 255.0
		}
		results[i] = vec
	}
	return results, nil
}

func newTestStore(t *testing.T) *Store {
	embedder := &mockEmbedder{dims: 8}
	store, err := NewStore(t.TempDir(), embedder)
	require.NoError(t, err)
	return store
}

// --- Cosine Similarity Tests ---

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float64{1, 2, 3}
	score, err := CosineSimilarity(a, a)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, score, 0.001)
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	score, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, score, 0.001)
}

func TestCosineSimilarity_DimensionMismatch(t *testing.T) {
	_, err := CosineSimilarity([]float64{1, 2}, []float64{1, 2, 3})
	assert.Error(t, err)
}

// --- Store Tests ---

func TestStore_Add(t *testing.T) {
	store := newTestStore(t)

	id, err := store.Add(context.Background(), "Hello world", "test", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, 1, store.Count())
}

func TestStore_Search(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	store.Add(ctx, "Go programming language", "docs", nil)
	store.Add(ctx, "Python programming language", "docs", nil)
	store.Add(ctx, "JavaScript web development", "docs", nil)

	results, err := store.Search(ctx, "Go programming", 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0].Rank)
	assert.Equal(t, 2, results[1].Rank)
	// First result should be most similar to "Go programming".
	assert.True(t, results[0].Score >= results[1].Score)
}

func TestStore_Remove(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	id, _ := store.Add(ctx, "test doc", "test", nil)
	assert.Equal(t, 1, store.Count())

	removed := store.Remove(id)
	assert.True(t, removed)
	assert.Equal(t, 0, store.Count())
}

func TestStore_Remove_NotFound(t *testing.T) {
	store := newTestStore(t)
	assert.False(t, store.Remove("nonexistent"))
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	embedder := &mockEmbedder{dims: 8}

	store1, err := NewStore(dir, embedder)
	require.NoError(t, err)
	store1.Add(context.Background(), "persistent doc", "test", nil)

	store2, err := NewStore(dir, embedder)
	require.NoError(t, err)
	assert.Equal(t, 1, store2.Count())
}

func TestStore_NoEmbedder(t *testing.T) {
	store, err := NewStore(t.TempDir(), nil)
	require.NoError(t, err)

	_, err = store.Add(context.Background(), "test", "test", nil)
	assert.Error(t, err)
}

func TestStore_AddChunked(t *testing.T) {
	store := newTestStore(t)

	longText := ""
	for i := 0; i < 200; i++ {
		longText += "Word "
	}

	ids, err := store.AddChunked(context.Background(), longText, "test", nil, ChunkOptions{
		MaxChunkSize: 100,
		Overlap:      20,
	})
	require.NoError(t, err)
	assert.True(t, len(ids) > 1)
	assert.Equal(t, len(ids), store.Count())
}

// --- Chunker Tests ---

func TestChunkText_Short(t *testing.T) {
	chunks := ChunkText("Hello world", DefaultChunkOptions())
	assert.Len(t, chunks, 1)
	assert.Equal(t, "Hello world", chunks[0])
}

func TestChunkText_Empty(t *testing.T) {
	chunks := ChunkText("", DefaultChunkOptions())
	assert.Nil(t, chunks)
}

func TestChunkText_Long(t *testing.T) {
	text := ""
	for i := 0; i < 100; i++ {
		text += "This is a test sentence. "
	}

	chunks := ChunkText(text, ChunkOptions{MaxChunkSize: 200, Overlap: 30})
	assert.True(t, len(chunks) > 1)

	// Each chunk should be <= max size.
	for _, c := range chunks {
		assert.True(t, len(c) <= 200, "chunk too long: %d", len(c))
	}
}

func TestChunkText_SentenceBoundary(t *testing.T) {
	text := "First sentence. Second sentence. Third sentence. Fourth sentence."
	chunks := ChunkText(text, ChunkOptions{MaxChunkSize: 40, Overlap: 5})
	assert.True(t, len(chunks) >= 2)
}

// --- OpenAI Embedder Tests ---

func TestOpenAIEmbedder_Defaults(t *testing.T) {
	e := NewOpenAIEmbedder(OpenAIEmbedderConfig{})
	assert.Equal(t, "openai", e.Name())
	assert.Equal(t, 1536, e.Dimensions())
}

func TestOpenAIEmbedder_CustomDims(t *testing.T) {
	e := NewOpenAIEmbedder(OpenAIEmbedderConfig{Dimensions: 384})
	assert.Equal(t, 384, e.Dimensions())
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 2, 3}
	score, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.True(t, math.IsNaN(score) || score == 0)
}

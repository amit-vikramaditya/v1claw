package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// Store is a local vector knowledge base backed by a JSON file.
// For lightweight deployments (V1Claw's philosophy), this avoids
// external dependencies like SQLite vector extensions or separate
// vector databases.
type Store struct {
	mu       sync.RWMutex
	path     string
	docs     []Document
	embedder EmbeddingProvider
}

// NewStore creates a new knowledge store at the given directory.
func NewStore(dir string, embedder EmbeddingProvider) (*Store, error) {
	path := filepath.Join(dir, "knowledge.json")
	s := &Store{
		path:     path,
		embedder: embedder,
	}
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("knowledge store load: %w", err)
	}
	logger.InfoCF("knowledge", "Knowledge store loaded", map[string]interface{}{
		"documents": len(s.docs), "path": path,
	})
	return s, nil
}

// Add indexes a document into the store.
func (s *Store) Add(ctx context.Context, content, source string, metadata map[string]string) (string, error) {
	if s.embedder == nil {
		return "", fmt.Errorf("no embedding provider configured")
	}

	embedding, err := s.embedder.Embed(ctx, content)
	if err != nil {
		return "", fmt.Errorf("embed: %w", err)
	}

	doc := Document{
		ID:        fmt.Sprintf("doc_%d", time.Now().UnixNano()),
		Content:   content,
		Source:    source,
		Metadata:  metadata,
		Embedding: embedding,
	}

	s.mu.Lock()
	s.docs = append(s.docs, doc)
	s.save()
	s.mu.Unlock()

	logger.DebugCF("knowledge", "Document added", map[string]interface{}{
		"id": doc.ID, "source": source, "content_len": len(content),
	})

	return doc.ID, nil
}

// AddChunked splits content into chunks and indexes each.
func (s *Store) AddChunked(ctx context.Context, content, source string, metadata map[string]string, opts ChunkOptions) ([]string, error) {
	chunks := ChunkText(content, opts)
	var ids []string

	for i, chunk := range chunks {
		meta := make(map[string]string)
		for k, v := range metadata {
			meta[k] = v
		}
		meta["chunk_index"] = fmt.Sprintf("%d", i)
		meta["chunk_total"] = fmt.Sprintf("%d", len(chunks))

		id, err := s.Add(ctx, chunk, source, meta)
		if err != nil {
			return ids, fmt.Errorf("add chunk %d: %w", i, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// Search finds the top-K most similar documents to the query.
func (s *Store) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("no embedding provider configured")
	}
	if topK <= 0 {
		topK = 5
	}

	queryEmbedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []SearchResult
	for _, doc := range s.docs {
		if len(doc.Embedding) == 0 {
			continue
		}
		score, err := CosineSimilarity(queryEmbedding, doc.Embedding)
		if err != nil {
			continue
		}
		results = append(results, SearchResult{
			Document: Document{
				ID:       doc.ID,
				Content:  doc.Content,
				Source:   doc.Source,
				Metadata: doc.Metadata,
			},
			Score: score,
		})
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	for i := range results {
		results[i].Rank = i + 1
	}

	return results, nil
}

// Remove deletes a document by ID.
func (s *Store) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, doc := range s.docs {
		if doc.ID == id {
			s.docs = append(s.docs[:i], s.docs[i+1:]...)
			s.save()
			return true
		}
	}
	return false
}

// Count returns the number of indexed documents.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}

// --- Persistence ---

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.docs = []Document{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.docs)
}

func (s *Store) save() {
	data, err := json.Marshal(s.docs)
	if err != nil {
		logger.ErrorC("knowledge", fmt.Sprintf("Failed to marshal knowledge store: %v", err))
		return
	}
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		logger.ErrorC("knowledge", fmt.Sprintf("Failed to write knowledge store: %v", err))
		return
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		logger.ErrorC("knowledge", fmt.Sprintf("Failed to rename knowledge store file: %v", err))
	}
}

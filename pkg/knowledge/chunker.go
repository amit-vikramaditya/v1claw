package knowledge

import (
	"strings"
	"unicode"
)

// ChunkText splits text into overlapping chunks suitable for embedding.
func ChunkText(text string, opts ChunkOptions) []string {
	if opts.MaxChunkSize <= 0 {
		opts.MaxChunkSize = 512
	}
	if opts.Overlap < 0 {
		opts.Overlap = 0
	}
	if opts.Overlap >= opts.MaxChunkSize {
		opts.Overlap = opts.MaxChunkSize / 4
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}
	if len(text) <= opts.MaxChunkSize {
		return []string{text}
	}

	var chunks []string
	step := opts.MaxChunkSize - opts.Overlap
	if step <= 0 {
		step = 1
	}

	for start := 0; start < len(text); start += step {
		end := start + opts.MaxChunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := text[start:end]

		// Try to break at a sentence or word boundary.
		if end < len(text) {
			if idx := lastSentenceBreak(chunk); idx > len(chunk)/2 {
				chunk = chunk[:idx+1]
			} else if idx := lastWordBreak(chunk); idx > len(chunk)/2 {
				chunk = chunk[:idx+1]
			}
		}

		chunk = strings.TrimSpace(chunk)
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// If we trimmed the chunk, adjust the next start.
		if len(chunk) < opts.MaxChunkSize && end < len(text) {
			start = start + len(chunk) - opts.Overlap - step + step
		}
	}

	return dedup(chunks)
}

func lastSentenceBreak(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' || s[i] == '!' || s[i] == '?' || s[i] == '\n' {
			return i
		}
	}
	return -1
}

func lastWordBreak(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if unicode.IsSpace(rune(s[i])) {
			return i
		}
	}
	return -1
}

func dedup(chunks []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, c := range chunks {
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}
	return result
}

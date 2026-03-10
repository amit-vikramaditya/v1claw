package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/providers"
)

type Session struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	storage  string
}

func NewSessionManager(storage string) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		storage:  storage,
	}

	if storage != "" {
		os.MkdirAll(storage, 0700)
		sm.loadSessions()
	}

	return sm
}

func (sm *SessionManager) GetOrCreate(key string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if ok {
		return session
	}

	session = &Session{
		Key:      key,
		Messages: []providers.Message{},
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	sm.sessions[key] = session

	return session
}

func (sm *SessionManager) AddMessage(sessionKey, role, content string) {
	sm.AddFullMessage(sessionKey, providers.Message{
		Role:    role,
		Content: content,
	})
}

// AddFullMessage adds a complete message with tool calls and tool call ID to the session.
// This is used to save the full conversation flow including tool calls and tool results.
func (sm *SessionManager) AddFullMessage(sessionKey string, msg providers.Message) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionKey]
	if !ok {
		session = &Session{
			Key:      sessionKey,
			Messages: []providers.Message{},
			Created:  time.Now(),
		}
		sm.sessions[sessionKey] = session
	}

	session.Messages = append(session.Messages, msg)
	session.Updated = time.Now()
}

func (sm *SessionManager) GetHistory(key string) []providers.Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[key]
	if !ok {
		return []providers.Message{}
	}

	return deepCopyMessages(session.Messages)
}

func (sm *SessionManager) GetSummary(key string) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[key]
	if !ok {
		return ""
	}
	return session.Summary
}

func (sm *SessionManager) SetSummary(key string, summary string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if ok {
		session.Summary = summary
		session.Updated = time.Now()
	}
}

func (sm *SessionManager) TruncateHistory(key string, keepLast int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return
	}

	if keepLast <= 0 {
		session.Messages = []providers.Message{}
		session.Updated = time.Now()
		return
	}

	if len(session.Messages) <= keepLast {
		return
	}

	session.Messages = session.Messages[len(session.Messages)-keepLast:]
	session.Updated = time.Now()
}

// sanitizeFilename converts a session key into a cross-platform safe filename.
// Session keys use "channel:chatID" (e.g. "telegram:123456") but ':' is the
// volume separator on Windows, so filepath.Base would misinterpret the key.
// We replace it with '_'. The original key is preserved inside the JSON file,
// so loadSessions still maps back to the right in-memory key.
func sanitizeFilename(key string) string {
	return strings.ReplaceAll(key, ":", "_")
}

func (sm *SessionManager) Save(key string) error {
	if sm.storage == "" {
		return nil
	}

	filename := sanitizeFilename(key)

	// filepath.IsLocal rejects empty names, "..", absolute paths, and
	// OS-reserved device names (NUL, COM1 … on Windows).
	// The extra checks reject "." and any directory separators so that
	// the session file is always written directly inside sm.storage.
	if filename == "." || !filepath.IsLocal(filename) || strings.ContainsAny(filename, `/\`) {
		return os.ErrInvalid
	}

	// Snapshot under read lock, then perform slow file I/O after unlock.
	sm.mu.RLock()
	stored, ok := sm.sessions[key]
	if !ok {
		sm.mu.RUnlock()
		return nil
	}

	snapshot := Session{
		Key:      stored.Key,
		Summary:  stored.Summary,
		Created:  stored.Created,
		Updated:  stored.Updated,
		Messages: deepCopyMessages(stored.Messages),
	}
	sm.mu.RUnlock()

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	sessionPath := filepath.Join(sm.storage, filename+".json")
	tmpFile, err := os.CreateTemp(sm.storage, "session-*.tmp")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Chmod(0600); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, sessionPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func (sm *SessionManager) loadSessions() error {
	files, err := os.ReadDir(sm.storage)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Skip oversized session files to prevent OOM on load.
		const maxSessionFileSize = 10 * 1024 * 1024 // 10 MB
		if info, err := file.Info(); err != nil || info.Size() > maxSessionFileSize {
			continue
		}

		sessionPath := filepath.Join(sm.storage, file.Name())
		data, err := os.ReadFile(sessionPath)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sm.sessions[session.Key] = &session
	}

	return nil
}

// SetHistory updates the messages of a session.
func (sm *SessionManager) SetHistory(key string, history []providers.Message) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if ok {
		session.Messages = deepCopyMessages(history)
		session.Updated = time.Now()
	}
}

// SummarizeAndTruncate atomically sets the session summary and truncates the
// message history in a single locked operation.  Using separate SetSummary and
// TruncateHistory calls introduces a race window where a new message arriving
// between the two steps could be incorrectly dropped.
func (sm *SessionManager) SummarizeAndTruncate(key string, summary string, keepLast int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[key]
	if !ok {
		return
	}

	session.Summary = summary
	session.Updated = time.Now()

	if keepLast <= 0 {
		session.Messages = []providers.Message{}
		return
	}

	if len(session.Messages) > keepLast {
		// Allocate a fresh slice so the old backing array can be GC'd.
		session.Messages = deepCopyMessages(session.Messages[len(session.Messages)-keepLast:])
	}
}

// deepCopyMessages returns a new slice where each Message and its pointer
// fields (ToolCall.Function *FunctionCall, ToolCall.Arguments map) are fully
// independent from the originals.  This prevents data races when multiple
// goroutines hold references to the same session history.
func deepCopyMessages(src []providers.Message) []providers.Message {
	if len(src) == 0 {
		return []providers.Message{}
	}
	dst := make([]providers.Message, len(src))
	for i, m := range src {
		dst[i] = providers.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if len(m.ToolCalls) > 0 {
			dst[i].ToolCalls = make([]providers.ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				dst[i].ToolCalls[j] = providers.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Name: tc.Name,
				}
				if tc.Function != nil {
					fn := *tc.Function // copy struct value
					dst[i].ToolCalls[j].Function = &fn
				}
				if tc.Arguments != nil {
					argsCopy := make(map[string]interface{}, len(tc.Arguments))
					for k, v := range tc.Arguments {
						argsCopy[k] = v
					}
					dst[i].ToolCalls[j].Arguments = argsCopy
				}
			}
		}
	}
	return dst
}

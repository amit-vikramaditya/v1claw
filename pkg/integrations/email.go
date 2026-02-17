package integrations

import (
	"context"
	"fmt"
	"time"
)

// EmailMessage represents an email.
type EmailMessage struct {
	ID      string    `json:"id"`
	From    string    `json:"from"`
	To      []string  `json:"to"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	IsHTML  bool      `json:"is_html,omitempty"`
	Date    time.Time `json:"date"`
	Read    bool      `json:"read"`
	Labels  []string  `json:"labels,omitempty"`
}

// EmailProvider is the interface for email backends.
type EmailProvider interface {
	// Fetch returns recent emails from the inbox.
	Fetch(ctx context.Context, limit int) ([]EmailMessage, error)
	// Send sends an email.
	Send(ctx context.Context, msg EmailMessage) error
	// MarkRead marks an email as read.
	MarkRead(ctx context.Context, id string) error
	// UnreadCount returns the number of unread emails.
	UnreadCount(ctx context.Context) (int, error)
	// Name returns the provider name.
	Name() string
}

// EmailConfig holds email integration configuration.
type EmailConfig struct {
	Enabled      bool   `json:"enabled"`
	IMAPServer   string `json:"imap_server"`   // e.g., "imap.gmail.com:993"
	SMTPServer   string `json:"smtp_server"`   // e.g., "smtp.gmail.com:587"
	Username     string `json:"username"`
	Password     string `json:"password"`
	PollInterval int    `json:"poll_interval"` // Minutes between inbox checks
}

// EmailManager provides a unified email interface with digest support.
type EmailManager struct {
	provider EmailProvider
	config   EmailConfig
}

// NewEmailManager creates an email manager.
func NewEmailManager(cfg EmailConfig, provider EmailProvider) *EmailManager {
	return &EmailManager{config: cfg, provider: provider}
}

// Digest returns a summary of recent unread emails.
func (m *EmailManager) Digest(ctx context.Context) (string, error) {
	count, err := m.provider.UnreadCount(ctx)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "📧 No unread emails.", nil
	}

	msgs, err := m.provider.Fetch(ctx, 5)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("📧 You have %d unread email(s):\n", count)
	for _, msg := range msgs {
		if !msg.Read {
			result += fmt.Sprintf("  • %s — %s\n", msg.From, msg.Subject)
		}
	}
	return result, nil
}

// QuickSend sends a simple text email.
func (m *EmailManager) QuickSend(ctx context.Context, to, subject, body string) error {
	return m.provider.Send(ctx, EmailMessage{
		To:      []string{to},
		Subject: subject,
		Body:    body,
	})
}

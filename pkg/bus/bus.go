package bus

import (
	"context"
	"sync"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

type MessageBus struct {
	inbound      chan InboundMessage
	outboundSubs []chan OutboundMessage
	handlers     map[string]MessageHandler
	mu           sync.RWMutex
	once         sync.Once
	done         chan struct{}
}

// Subscription represents an active outbound subscription.
// Call Unsubscribe when the subscriber is done to free resources.
type Subscription struct {
	C  <-chan OutboundMessage
	ch chan OutboundMessage
	mb *MessageBus
}

// Unsubscribe removes this subscription from the bus and closes the channel.
// It is safe to call multiple times.
func (s *Subscription) Unsubscribe() {
	s.mb.mu.Lock()
	defer s.mb.mu.Unlock()

	subs := s.mb.outboundSubs
	for i, c := range subs {
		if c == s.ch {
			s.mb.outboundSubs = append(subs[:i], subs[i+1:]...)
			close(s.ch)
			break
		}
	}
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inbound:      make(chan InboundMessage, 1024),
		outboundSubs: make([]chan OutboundMessage, 0),
		handlers:     make(map[string]MessageHandler),
		done:         make(chan struct{}),
	}
}

// PublishInbound delivers a message to the agent loop.
// Returns true if the message was queued, false if the buffer was full (dropped).
func (mb *MessageBus) PublishInbound(msg InboundMessage) bool {
	select {
	case <-mb.done:
		return false
	case mb.inbound <- msg:
		return true
	default:
		logger.WarnC("bus", "Inbound buffer full, dropping message to prevent deadlock")
		return false
	}
}

func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg, ok := <-mb.inbound:
		return msg, ok
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	select {
	case <-mb.done:
		return
	default:
	}
	for _, sub := range mb.outboundSubs {
		select {
		case sub <- msg:
		default:
			logger.WarnC("bus", "Outbound buffer full, dropping message to prevent deadlock for slow subscriber")
		}
	}
}

// SubscribeOutbound returns a Subscription whose C field receives outbound messages.
// Call Subscription.Unsubscribe() when done to prevent goroutine/memory leaks.
func (mb *MessageBus) SubscribeOutbound() *Subscription {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	ch := make(chan OutboundMessage, 1024)
	mb.outboundSubs = append(mb.outboundSubs, ch)
	return &Subscription{C: ch, ch: ch, mb: mb}
}

func (mb *MessageBus) RegisterHandler(channel string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[channel] = handler
}

func (mb *MessageBus) GetHandler(channel string) (MessageHandler, bool) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	handler, ok := mb.handlers[channel]
	return handler, ok
}

func (mb *MessageBus) Close() {
	mb.once.Do(func() {
		close(mb.done)
		// We don't close inbound/outbound channels here to allow
		// pending goroutines to finish without panicking on send.
	})
}

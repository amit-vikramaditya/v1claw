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

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inbound:      make(chan InboundMessage, 1024), // Increased buffer
		outboundSubs: make([]chan OutboundMessage, 0),
		handlers:     make(map[string]MessageHandler),
		done:         make(chan struct{}),
	}
}

func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	select {
	case <-mb.done:
		return
	case mb.inbound <- msg:
	default:
		logger.WarnC("bus", "Inbound buffer full, dropping message to prevent deadlock")
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

func (mb *MessageBus) SubscribeOutbound() <-chan OutboundMessage {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	ch := make(chan OutboundMessage, 1024)
	mb.outboundSubs = append(mb.outboundSubs, ch)
	return ch
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

package event

import (
	"context"
	"sync"
)

// MemoryBus provides an in-memory implementation of the event bus
type MemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewMemoryBus creates a new memory event bus
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]Handler),
	}
}

// Publish publishes an event to all registered handlers
func (b *MemoryBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			// Log error but continue processing other handlers
			// In a production system, we might want to handle this differently
			continue
		}
	}

	return nil
}

// Subscribe registers a handler for an event type
func (b *MemoryBus) Subscribe(eventType string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.handlers[eventType] == nil {
		b.handlers[eventType] = make([]Handler, 0)
	}
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes a handler for an event type
func (b *MemoryBus) Unsubscribe(eventType string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, ok := b.handlers[eventType]; ok {
		for i, h := range handlers {
			if h == handler {
				b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
	return nil
}

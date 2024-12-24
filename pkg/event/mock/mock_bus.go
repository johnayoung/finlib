package mock

import (
	"context"

	"github.com/johnayoung/finlib/pkg/event"
	"github.com/stretchr/testify/mock"
)

// MockBus is a mock implementation of the event.Bus interface
type MockBus struct {
	mock.Mock
}

// Publish implements event.Bus
func (m *MockBus) Publish(ctx context.Context, e event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

// Subscribe implements event.Bus
func (m *MockBus) Subscribe(eventType string, handler event.Handler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

// Unsubscribe implements event.Bus
func (m *MockBus) Unsubscribe(eventType string, handler event.Handler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

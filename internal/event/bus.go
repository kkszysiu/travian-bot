package event

import (
	"context"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Bus is a channel-based event bus that bridges Go events to the Wails frontend.
type Bus struct {
	ctx        context.Context
	mu         sync.RWMutex
	handlers   map[string][]func(interface{})
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]func(interface{})),
	}
}

// SetContext sets the Wails runtime context for frontend event emission.
func (b *Bus) SetContext(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ctx = ctx
}

// Emit publishes an event to both Go subscribers and the Wails frontend.
func (b *Bus) Emit(event string, data interface{}) {
	b.mu.RLock()
	ctx := b.ctx
	handlers := b.handlers[event]
	b.mu.RUnlock()

	// Notify Go subscribers
	for _, h := range handlers {
		h(data)
	}

	// Notify Wails frontend
	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, event, data)
	}
}

// On registers a Go-side handler for an event.
func (b *Bus) On(event string, handler func(interface{})) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[event] = append(b.handlers[event], handler)
}

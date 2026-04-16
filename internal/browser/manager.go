package browser

import (
	"fmt"
	"log/slog"
	"sync"
)

// Manager maintains a per-account browser pool.
type Manager struct {
	mu       sync.RWMutex
	browsers map[int]*Browser // accountID -> browser
	logger   *slog.Logger
}

// NewManager creates a browser manager.
func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		browsers: make(map[int]*Browser),
		logger:   logger,
	}
}

// Get returns the browser for the given account, or nil.
func (m *Manager) Get(accountID int) *Browser {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.browsers[accountID]
}

// Create launches a new browser for the account with the given config.
func (m *Manager) Create(accountID int, cfg Config) (*Browser, error) {
	// Close existing if any
	m.Close(accountID)

	b, err := New(cfg)
	if err != nil {
		return nil, fmt.Errorf("create browser for account %d: %w", accountID, err)
	}

	m.mu.Lock()
	m.browsers[accountID] = b
	m.mu.Unlock()

	m.logger.Info("browser created", "accountId", accountID)
	return b, nil
}

// Close shuts down the browser for an account.
func (m *Manager) Close(accountID int) {
	m.mu.Lock()
	b, ok := m.browsers[accountID]
	if ok {
		delete(m.browsers, accountID)
	}
	m.mu.Unlock()

	if b != nil {
		b.Close()
		m.logger.Info("browser closed", "accountId", accountID)
	}
}

// Shutdown closes all browsers.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	browsers := m.browsers
	m.browsers = make(map[int]*Browser)
	m.mu.Unlock()

	for id, b := range browsers {
		b.Close()
		m.logger.Info("browser closed", "accountId", id)
	}
}

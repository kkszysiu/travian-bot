package app

import (
	"sync"
	"time"
)

const maxLogEntries = 5000

type logEntry struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	Time    string `json:"time"`
}

type logStore struct {
	mu      sync.RWMutex
	entries map[int][]logEntry // accountID -> entries (ring buffer)
}

func newLogStore() *logStore {
	return &logStore{
		entries: make(map[int][]logEntry),
	}
}

func (ls *logStore) Add(accountID int, level, message string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	entry := logEntry{
		Message: message,
		Level:   level,
		Time:    time.Now().Format("15:04:05"),
	}

	entries := ls.entries[accountID]
	entries = append(entries, entry)
	if len(entries) > maxLogEntries {
		entries = entries[len(entries)-maxLogEntries:]
	}
	ls.entries[accountID] = entries
}

func (ls *logStore) Get(accountID int) []logEntry {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	entries := ls.entries[accountID]
	if entries == nil {
		return nil
	}
	// Return a copy
	result := make([]logEntry, len(entries))
	copy(result, entries)
	return result
}

func (ls *logStore) Clear(accountID int) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	delete(ls.entries, accountID)
}

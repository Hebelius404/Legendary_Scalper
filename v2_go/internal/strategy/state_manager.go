package strategy

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// StateManager handles atomic saving/loading of bot state
type StateManager struct {
	mu           sync.RWMutex
	FilePath     string
	ActiveTrades map[string]*MartingalePosition
}

// NewStateManager creates a manager pointing to the file
func NewStateManager(path string) *StateManager {
	return &StateManager{
		FilePath:     path,
		ActiveTrades: make(map[string]*MartingalePosition),
	}
}

// LoadState reads from JSON file
func (sm *StateManager) LoadState() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.FilePath)
	if os.IsNotExist(err) {
		return nil // No state file yet, clean start
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &sm.ActiveTrades); err != nil {
		return fmt.Errorf("corrupt state file: %v", err)
	}

	return nil
}

// SaveState writes to JSON file atomically
func (sm *StateManager) SaveState() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := json.MarshalIndent(sm.ActiveTrades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.FilePath, data, 0644)
}

// UpdatePosition updates a single position and saves state
func (sm *StateManager) UpdatePosition(pos *MartingalePosition) error {
	sm.mu.Lock()
	sm.ActiveTrades[pos.Symbol] = pos
	sm.mu.Unlock()
	return sm.SaveState()
}

// ClearPosition removes a position (after close) and saves
func (sm *StateManager) ClearPosition(symbol string) error {
	sm.mu.Lock()
	delete(sm.ActiveTrades, symbol)
	sm.mu.Unlock()
	return sm.SaveState()
}

// RemovePosition is an alias for ClearPosition to satisfy engine requirements
func (sm *StateManager) RemovePosition(symbol string) error {
	return sm.ClearPosition(symbol)
}

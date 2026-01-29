package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wisp-trading/wisp/pkg/live"
)

type fileStateStore struct {
	mu   sync.RWMutex
	path string
}

// NewFileStateStore creates a new file-based state store at ~/.wisp/.instances.json
func NewFileStateStore() (live.StateStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	wispDir := filepath.Join(homeDir, ".wisp")
	if err := os.MkdirAll(wispDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create wisp directory: %w", err)
	}

	stateFile := filepath.Join(wispDir, ".instances.json")

	return &fileStateStore{
		path: stateFile,
	}, nil
}

// Load reads persisted state from disk
func (fss *fileStateStore) Load() ([]*live.Instance, error) {
	fss.mu.RLock()
	defer fss.mu.RUnlock()

	data, err := os.ReadFile(fss.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*live.Instance{}, nil
		}
		return nil, err
	}

	var state struct {
		Instances []*live.Instance `json:"instances"`
		LastSaved time.Time        `json:"last_saved"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state.Instances, nil
}

// Save writes current state to disk (atomic)
func (fss *fileStateStore) Save(instances []*live.Instance) error {
	fss.mu.Lock()
	defer fss.mu.Unlock()

	state := struct {
		Instances []*live.Instance `json:"instances"`
		LastSaved time.Time        `json:"last_saved"`
	}{
		Instances: instances,
		LastSaved: time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temp file first (atomic write)
	tmpPath := fss.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, fss.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// GetPath returns the path to the state file
func (fss *fileStateStore) GetPath() string {
	return fss.path
}

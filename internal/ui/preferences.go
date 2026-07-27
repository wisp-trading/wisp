package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Preferences stores user UI preferences
type Preferences struct {
	Theme string `yaml:"theme"`
}

// getPreferencesPath returns the path to the preferences file
func getPreferencesPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	wispDir := filepath.Join(home, ".wisp")

	// Ensure directory exists
	if err := os.MkdirAll(wispDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .wisp directory: %w", err)
	}

	return filepath.Join(wispDir, "preferences.yml"), nil
}

// LoadPreferences loads user preferences from disk
func LoadPreferences() (*Preferences, error) {
	path, err := getPreferencesPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Preferences{Theme: "default"}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read preferences: %w", err)
	}

	var prefs Preferences
	if err := yaml.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to parse preferences: %w", err)
	}

	// Default to "default" if theme not set
	if prefs.Theme == "" {
		prefs.Theme = "default"
	}

	return &prefs, nil
}

// SavePreferences saves user preferences to disk
func SavePreferences(prefs *Preferences) error {
	path, err := getPreferencesPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences: %w", err)
	}

	return nil
}

// LoadThemeFromPreferences loads and applies the theme from saved preferences.
// Missing preference files are treated as success (default theme remains active).
func LoadThemeFromPreferences() error {
	prefs, err := LoadPreferences()
	if err != nil {
		// First run / unreadable prefs — default theme is intentional.
		prefs = &Preferences{Theme: "default"}
	}

	if err := SetTheme(prefs.Theme); err != nil {
		return SetTheme("default")
	}
	return nil
}

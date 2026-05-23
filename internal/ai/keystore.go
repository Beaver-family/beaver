package ai

import (
	"os"
	"path/filepath"
	"strings"
)

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "beaver", "apikey")
}

// LoadSavedKey reads the API key from ~/.config/beaver/apikey.
func LoadSavedKey() string {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SaveKey writes the API key to ~/.config/beaver/apikey (mode 0600).
func SaveKey(key string) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(key)), 0600)
}

func modelPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "beaver", "model")
}

// LoadSavedModel returns the saved model ID, or DefaultModel if none saved.
func LoadSavedModel() string {
	data, err := os.ReadFile(modelPath())
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return DefaultModel
	}
	return strings.TrimSpace(string(data))
}

// SaveModel persists the chosen model ID.
func SaveModel(modelID string) error {
	path := modelPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(modelID), 0644)
}

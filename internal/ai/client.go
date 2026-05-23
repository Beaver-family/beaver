package ai

import (
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// NewClient returns an Anthropic client using the given API key.
func NewClient(apiKey string) *anthropic.Client {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &c
}

// ResolveAPIKey returns the API key from env var or saved config.
// Returns empty string if not found anywhere.
func ResolveAPIKey() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	return LoadSavedKey()
}

package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func NewOpenAIClient(apiKey string) *openai.Client {
	c := openai.NewClient(option.WithAPIKey(apiKey))
	return &c
}

func openaiKeyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "beaver", "openai_apikey")
}

func LoadSavedOpenAIKey() string {
	data, err := os.ReadFile(openaiKeyPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SaveOpenAIKey(key string) error {
	path := openaiKeyPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(key)), 0600)
}

func ResolveOpenAIKey() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	return LoadSavedOpenAIKey()
}

func ValidateOpenAIKey(apiKey string) error {
	client := NewOpenAIClient(apiKey)
	_, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model:           "gpt-4.1-mini",
		Messages:        []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")},
		MaxCompletionTokens: openai.Int(1),
	})
	return err
}

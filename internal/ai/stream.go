package ai

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
)

// ModelOption describes one entry in the model picker.
type ModelOption struct {
	ID          string
	Name        string
	Description string
	Provider    string
}

var Models = []ModelOption{
	// ── Anthropic ──────────────────────────────────────────────────────────
	{
		ID:          "claude-opus-4-7",
		Name:        "Claude Opus 4.7",
		Description: "Most capable · best for complex tasks · Jan 2026 knowledge",
		Provider:    ProviderAnthropic,
	},
	{
		ID:          "claude-sonnet-4-6",
		Name:        "Claude Sonnet 4.6",
		Description: "Recommended · fast & smart · Aug 2025 knowledge",
		Provider:    ProviderAnthropic,
	},
	{
		ID:          "claude-haiku-4-5-20251001",
		Name:        "Claude Haiku 4.5",
		Description: "Fastest & cheapest · great for quick questions",
		Provider:    ProviderAnthropic,
	},
	// ── OpenAI ─────────────────────────────────────────────────────────────
	{
		ID:          "gpt-5.1-codex",
		Name:        "GPT Codex (5.1)",
		Description: "OpenAI's latest coding model · best for software engineering",
		Provider:    ProviderOpenAI,
	},
	{
		ID:          "gpt-4.1",
		Name:        "GPT-4.1",
		Description: "OpenAI's latest GPT-4 model · fast & capable",
		Provider:    ProviderOpenAI,
	},
	{
		ID:          "o3",
		Name:        "o3",
		Description: "OpenAI's frontier reasoning model · excels at hard problems",
		Provider:    ProviderOpenAI,
	},
}

const DefaultModel       = "claude-sonnet-4-6"
const DefaultOpenAIModel = "gpt-5.1-codex"

// ProviderForModel returns the provider string for the given model ID.
func ProviderForModel(id string) string {
	for _, m := range Models {
		if m.ID == id {
			return m.Provider
		}
	}
	return ProviderAnthropic
}

// Message is a single turn in the conversation.
// Role is "user", "assistant", or "tool" (UI-only display; skipped when sending to API).
type Message struct {
	Role    string
	Content string
}

// StreamTokenMsg carries one streamed text chunk.
type StreamTokenMsg struct{ Text string }

// StreamDoneMsg signals the stream has finished.
type StreamDoneMsg struct{}

// ValidateKey sends a minimal request to verify the key works.
func ValidateKey(apiKey string) error {
	client := NewClient(apiKey)
	_, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.Model(DefaultModel),
		MaxTokens: 1,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("hi")),
		},
	})
	return err
}

// StreamStartMsg carries the agent event channel back to the model.
type StreamStartMsg struct{ Ch <-chan AgentEvent }

// ReadAgentEvent reads one event from the channel and maps it to a tea.Msg.
// Re-schedule this after each non-terminal message.
func ReadAgentEvent(ch <-chan AgentEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return StreamDoneMsg{}
		}
		switch ev := event.(type) {
		case AgentTextToken:
			return StreamTokenMsg{Text: ev.Text}
		case AgentToolStarted:
			return ev
		case AgentToolDone:
			return ev
		}
		return StreamDoneMsg{}
	}
}


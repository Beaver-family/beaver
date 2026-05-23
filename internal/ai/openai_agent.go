package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenAIAgentCmd starts the OpenAI agentic loop and returns StreamStartMsg.
func OpenAIAgentCmd(client *openai.Client, history []Message, systemPrompt, modelID, rootPath string) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan AgentEvent, 256)
		go runOpenAIAgentLoop(client, history, systemPrompt, modelID, rootPath, ch)
		return StreamStartMsg{Ch: ch}
	}
}

func runOpenAIAgentLoop(
	client *openai.Client,
	history []Message,
	systemPrompt, modelID, rootPath string,
	ch chan<- AgentEvent,
) {
	defer close(ch)

	if modelID == "" {
		modelID = DefaultOpenAIModel
	}

	// build initial messages
	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
	if systemPrompt != "" {
		msgs = append(msgs, openai.SystemMessage(systemPrompt))
	}
	for _, m := range history {
		switch m.Role {
		case "user":
			msgs = append(msgs, openai.UserMessage(m.Content))
		case "assistant":
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		}
	}

	for {
		stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
			Model:    modelID,
			Messages: msgs,
			Tools:    BuildOpenAITools(),
		})

		// accumulate tool calls by index while streaming text tokens live
		type pendingCall struct {
			id      string
			name    string
			argsBuf strings.Builder
		}
		pending := map[int64]*pendingCall{}
		announcedTools := map[string]bool{}
		var assistantText strings.Builder

		for stream.Next() {
			chunk := stream.Current()
			for _, choice := range chunk.Choices {
				delta := choice.Delta
				if delta.Content != "" {
					assistantText.WriteString(delta.Content)
					ch <- AgentTextToken{Text: delta.Content}
				}
				for _, tc := range delta.ToolCalls {
					idx := tc.Index
					if _, exists := pending[idx]; !exists {
						pending[idx] = &pendingCall{}
					}
					p := pending[idx]
					if tc.ID != "" {
						p.id = tc.ID
					}
					if tc.Function.Name != "" {
						p.name = tc.Function.Name
						if !announcedTools[p.name] {
							announcedTools[p.name] = true
							ch <- AgentToolStarted{Name: p.name}
						}
					}
					p.argsBuf.WriteString(tc.Function.Arguments)
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- AgentTextToken{Text: fmt.Sprintf("\n\n[error: %s]", err.Error())}
			return
		}

		// no tool calls → done
		if len(pending) == 0 {
			return
		}

		// build assistant message with tool calls
		assistantMsg := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{},
		}
		if t := assistantText.String(); t != "" {
			assistantMsg.OfAssistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(t),
			}
		}
		// attach tool calls in index order
		for i := int64(0); ; i++ {
			p, ok := pending[i]
			if !ok {
				break
			}
			assistantMsg.OfAssistant.ToolCalls = append(
				assistantMsg.OfAssistant.ToolCalls,
				openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: p.id,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      p.name,
							Arguments: p.argsBuf.String(),
						},
					},
				},
			)
		}
		msgs = append(msgs, assistantMsg)

		// execute tools and collect results
		for i := int64(0); ; i++ {
			p, ok := pending[i]
			if !ok {
				break
			}
			result, isError := ExecuteTool(p.name, json.RawMessage(p.argsBuf.String()), rootPath)
			summary := result
			if len(summary) > 80 {
				summary = summary[:80] + "..."
			}
			ch <- AgentToolDone{Name: p.name, Summary: summary, IsError: isError}
			msgs = append(msgs, openai.ToolMessage(result, p.id))
		}
	}
}

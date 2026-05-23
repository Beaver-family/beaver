package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

// AgentEvent is sent through the streaming channel.
type AgentEvent interface{ agentEvent() }

type AgentTextToken struct{ Text string }
type AgentToolStarted struct{ Name string }
type AgentToolDone struct {
	Name    string
	Summary string
	IsError bool
}

func (AgentTextToken) agentEvent()  {}
func (AgentToolStarted) agentEvent() {}
func (AgentToolDone) agentEvent()   {}

// AgentCmd launches the full agentic loop in a goroutine and returns StreamStartMsg.
func AgentCmd(client *anthropic.Client, history []Message, systemPrompt, modelID, rootPath string) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan AgentEvent, 256)
		go runAgentLoop(client, history, systemPrompt, modelID, rootPath, ch)
		return StreamStartMsg{Ch: ch}
	}
}

func runAgentLoop(
	client *anthropic.Client,
	history []Message,
	systemPrompt, modelID, rootPath string,
	ch chan<- AgentEvent,
) {
	defer close(ch)

	if modelID == "" {
		modelID = DefaultModel
	}

	// build initial message params from history; skip "tool" UI-display messages
	msgs := make([]anthropic.MessageParam, 0, len(history))
	for _, m := range history {
		switch m.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	for {
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(modelID),
			MaxTokens: 4096,
			Tools:     BuildTools(),
			Messages:  msgs,
		}
		if systemPrompt != "" {
			params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
		}

		stream := client.Messages.NewStreaming(context.Background(), params)
		accumulated := anthropic.Message{}

		// track tool starts announced during stream so we don't double-announce
		announcedTools := map[string]bool{}

		for stream.Next() {
			event := stream.Current()
			_ = accumulated.Accumulate(event)

			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				if tu, ok := ev.ContentBlock.AsAny().(anthropic.ToolUseBlock); ok {
					if !announcedTools[tu.ID] {
						announcedTools[tu.ID] = true
						ch <- AgentToolStarted{Name: tu.Name}
					}
				}
			case anthropic.ContentBlockDeltaEvent:
				if delta, ok := ev.Delta.AsAny().(anthropic.TextDelta); ok && delta.Text != "" {
					ch <- AgentTextToken{Text: delta.Text}
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- AgentTextToken{Text: fmt.Sprintf("\n\n[error: %s]", err.Error())}
			return
		}

		// collect tool_use blocks from accumulated response
		var toolUses []anthropic.ToolUseBlock
		for _, block := range accumulated.Content {
			if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				toolUses = append(toolUses, tu)
			}
		}

		if len(toolUses) == 0 {
			return // final text response — done
		}

		// rebuild assistant message param from accumulated content
		var assistantBlocks []anthropic.ContentBlockParamUnion
		for _, block := range accumulated.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(b.Text))
			case anthropic.ToolUseBlock:
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(b.ID, b.Input, b.Name))
			}
		}
		msgs = append(msgs, anthropic.NewAssistantMessage(assistantBlocks...))

		// execute tools and collect results
		var toolResults []anthropic.ContentBlockParamUnion
		for _, tu := range toolUses {
			result, isError := ExecuteTool(tu.Name, json.RawMessage(tu.Input), rootPath)

			summary := result
			if len(summary) > 80 {
				summary = summary[:80] + "..."
			}
			ch <- AgentToolDone{Name: tu.Name, Summary: summary, IsError: isError}

			toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.ID, result, isError))
		}

		// add tool results and loop for next turn
		msgs = append(msgs, anthropic.NewUserMessage(toolResults...))
	}
}

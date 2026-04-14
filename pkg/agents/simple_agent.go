package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"GoAgent/pkg/core"
	"GoAgent/pkg/tools"
)

type SimpleAgent struct {
	*core.BaseAgent
	ToolRegistry   *tools.ToolRegistry
	EnableToolCall bool
	MaxIterations  int
}

type SimpleAgentOption func(*SimpleAgent)

func WithToolRegistry(reg *tools.ToolRegistry) SimpleAgentOption {
	return func(a *SimpleAgent) { a.ToolRegistry = reg }
}

func WithToolCall(enable bool) SimpleAgentOption {
	return func(a *SimpleAgent) { a.EnableToolCall = enable }
}

func WithSimpleMaxIterations(max int) SimpleAgentOption {
	return func(a *SimpleAgent) { a.MaxIterations = max }
}

func NewSimpleAgent(name string, llm *core.LLM, systemPrompt string, opts ...SimpleAgentOption) *SimpleAgent {
	agent := &SimpleAgent{
		BaseAgent:      core.NewBaseAgent(name, llm, systemPrompt, nil),
		ToolRegistry:   nil,
		EnableToolCall: false,
		MaxIterations:  3,
	}

	for _, opt := range opts {
		opt(agent)
	}

	if agent.ToolRegistry != nil {
		agent.EnableToolCall = true
	}

	return agent
}

func (a *SimpleAgent) GetEnhancedSystemPrompt() string {
	prompt := a.SystemPrompt
	if prompt == "" {
		prompt = "You are a helpful AI assistant."
	}

	if !a.EnableToolCall || a.ToolRegistry == nil {
		return prompt
	}

	toolsDesc := a.ToolRegistry.GetDescription()
	if toolsDesc == "No tools available" {
		return prompt
	}

	toolsSection := "\n\n## Available Tools\n" +
		"You can use the following tools to help answer questions:\n" +
		toolsDesc + "\n\n" +
		"## Tool Call Format\n" +
		"When you need to use a tool, use the following format:\n" +
		"`[TOOL_CALL:{tool_name}:{parameters}]`\n\n" +
		"### Parameter Format\n" +
		"1. Multiple parameters: use `key=value` format, separated by commas\n" +
		"   Example: `[TOOL_CALL:calculator:a=12,b=8]`\n" +
		"2. Single parameter: use `key=value`\n" +
		"   Example: `[TOOL_CALL:search:query=Python]`\n\n" +
		"### Important Notes\n" +
		"- Parameter names must match the tool definition exactly\n" +
		"- Tool results will be automatically inserted into the conversation\n"

	return prompt + toolsSection
}

func (a *SimpleAgent) parseToolCalls(text string) []struct {
	ToolName   string
	Parameters string
	Original   string
} {
	pattern := regexp.MustCompile(`\[TOOL_CALL:([^:]+):([^\]]+)\]`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	var calls []struct {
		ToolName   string
		Parameters string
		Original   string
	}

	for _, match := range matches {
		if len(match) >= 3 {
			calls = append(calls, struct {
				ToolName   string
				Parameters string
				Original   string
			}{
				ToolName:   strings.TrimSpace(match[1]),
				Parameters: strings.TrimSpace(match[2]),
				Original:   match[0],
			})
		}
	}
	return calls
}

func (a *SimpleAgent) executeToolCall(toolName, params string) string {
	if a.ToolRegistry == nil {
		return "❌ Error: Tool registry not configured"
	}

	tool := a.ToolRegistry.Get(toolName)
	if tool == nil {
		return fmt.Sprintf("❌ Error: Tool '%s' not found", toolName)
	}

	paramMap := a.parseToolParams(toolName, params)
	result, err := tool.Run(paramMap)
	if err != nil {
		return fmt.Sprintf("❌ Tool execution failed: %v", err)
	}
	return fmt.Sprintf("🔧 Tool %s result:\n%s", toolName, result)
}

func (a *SimpleAgent) parseToolParams(toolName, params string) map[string]interface{} {
	result := make(map[string]interface{})
	params = strings.TrimSpace(params)

	if params == "" {
		return result
	}

	if strings.HasPrefix(params, "{") && strings.HasSuffix(params, "}") {
		result["input"] = params
		return result
	}

	if strings.Contains(params, "=") {
		pairs := strings.Split(params, ",")
		for _, pair := range pairs {
			if strings.Contains(pair, "=") {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	} else {
		result["input"] = params
	}

	return result
}

func (a *SimpleAgent) Run(ctx context.Context, input string) (string, error) {
	history := a.GetHistory()
	messages := make([]core.ChatMessage, 0, len(history)+4)

	enhancedPrompt := a.GetEnhancedSystemPrompt()
	messages = append(messages, core.ChatMessage{Role: "system", Content: enhancedPrompt})

	for _, msg := range history {
		messages = append(messages, core.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	messages = append(messages, core.ChatMessage{Role: "user", Content: input})

	if !a.EnableToolCall {
		response, err := a.LLM.Invoke(ctx, messages)
		if err != nil {
			return "", err
		}
		a.AddMessage(core.NewUserMessage(input))
		a.AddMessage(core.NewAssistantMessage(response))
		return response, nil
	}

	currentIteration := 0
	finalResponse := ""

	for currentIteration < a.MaxIterations {
		response, err := a.LLM.Invoke(ctx, messages)
		if err != nil {
			return "", err
		}

		calls := a.parseToolCalls(response)

		if len(calls) == 0 {
			finalResponse = response
			break
		}

		var toolResults []string
		cleanResponse := response

		for _, call := range calls {
			result := a.executeToolCall(call.ToolName, call.Parameters)
			toolResults = append(toolResults, result)
			cleanResponse = strings.ReplaceAll(cleanResponse, call.Original, "")
		}

		messages = append(messages, core.ChatMessage{Role: "assistant", Content: cleanResponse})
		messages = append(messages, core.ChatMessage{
			Role:    "user",
			Content: "Tool results:\n" + strings.Join(toolResults, "\n\n") + "\n\nPlease provide the complete answer based on these results.",
		})

		currentIteration++
	}

	if currentIteration >= a.MaxIterations && finalResponse == "" {
		response, err := a.LLM.Invoke(ctx, messages)
		if err != nil {
			return "", err
		}
		finalResponse = response
	}

	a.AddMessage(core.NewUserMessage(input))
	a.AddMessage(core.NewAssistantMessage(finalResponse))

	return finalResponse, nil
}

func (a *SimpleAgent) Stream(ctx context.Context, input string) (<-chan string, <-chan error) {
	ch := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		messages := []core.ChatMessage{
			{Role: "system", Content: a.GetEnhancedSystemPrompt()},
			{Role: "user", Content: input},
		}

		respCh, respErrCh := a.LLM.Stream(ctx, messages)

		select {
		case err := <-respErrCh:
			errCh <- err
			return
		default:
		}

		var fullResp strings.Builder
		for {
			select {
			case chunk, ok := <-respCh:
				if !ok {
					goto done
				}
				ch <- chunk
				fullResp.WriteString(chunk)
			case err := <-respErrCh:
				errCh <- err
				return
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}

	done:
		a.AddMessage(core.NewUserMessage(input))
		a.AddMessage(core.NewAssistantMessage(fullResp.String()))
	}()

	return ch, errCh
}

func (a *SimpleAgent) AddTool(tool tools.Tool) {
	if a.ToolRegistry == nil {
		a.ToolRegistry = tools.NewToolRegistry()
		a.EnableToolCall = true
	}
	a.ToolRegistry.Register(tool)
}

func (a *SimpleAgent) RemoveTool(name string) bool {
	if a.ToolRegistry != nil {
		return a.ToolRegistry.Unregister(name)
	}
	return false
}

func (a *SimpleAgent) ListTools() []string {
	if a.ToolRegistry != nil {
		return a.ToolRegistry.List()
	}
	return nil
}

func (a *SimpleAgent) HasTools() bool {
	return a.EnableToolCall && a.ToolRegistry != nil
}

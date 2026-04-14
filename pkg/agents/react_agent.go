package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"GoAgent/pkg/core"
	"GoAgent/pkg/tools"
)

const DefaultReActPrompt = `You are an AI assistant with reasoning and acting capabilities. You can analyze problems and call appropriate tools to retrieve information, ultimately providing accurate answers.

## Available Tools
{tools}

## Workflow
Please respond in the following format, one step at a time:

**Thought:** Analyze the current problem, think about what information is needed or what action to take.
**Action:** Choose an action, format must be one of:
- tool_name[tool_input] - Call specified tool
- Finish[final_answer] - When you have enough information to give the final answer

## Important Notes
1. Each response must contain both Thought and Action
2. Tool call format must strictly follow: tool_name[parameters]
3. Only use Finish when you are confident you have enough information to answer
4. If the tool returns insufficient information, continue using other tools

## Current Task
**Question:** {question}

## Execution History
{history}

Now start your reasoning and action:`

type ReActAgent struct {
	*core.BaseAgent
	ToolRegistry   *tools.ToolRegistry
	MaxSteps       int
	PromptTemplate string
	CurrentHistory []string
}

type ReActAgentOption func(*ReActAgent)

func WithMaxSteps(max int) ReActAgentOption {
	return func(a *ReActAgent) { a.MaxSteps = max }
}

func WithReActPrompt(template string) ReActAgentOption {
	return func(a *ReActAgent) { a.PromptTemplate = template }
}

func NewReActAgent(name string, llm *core.LLM, reg *tools.ToolRegistry, systemPrompt string, opts ...ReActAgentOption) *ReActAgent {
	agent := &ReActAgent{
		BaseAgent:      core.NewBaseAgent(name, llm, systemPrompt, nil),
		ToolRegistry:   reg,
		MaxSteps:       5,
		PromptTemplate: DefaultReActPrompt,
		CurrentHistory: make([]string, 0),
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func (a *ReActAgent) buildPrompt(question string) string {
	toolsDesc := "No tools available"
	if a.ToolRegistry != nil {
		toolsDesc = a.ToolRegistry.GetDescription()
	}

	historyStr := ""
	if len(a.CurrentHistory) > 0 {
		historyStr = strings.Join(a.CurrentHistory, "\n")
	}

	prompt := a.PromptTemplate
	prompt = strings.ReplaceAll(prompt, "{tools}", toolsDesc)
	prompt = strings.ReplaceAll(prompt, "{question}", question)
	prompt = strings.ReplaceAll(prompt, "{history}", historyStr)

	return prompt
}

func (a *ReActAgent) parseOutput(text string) (thought, action string) {
	thoughtPattern := regexp.MustCompile(`(?is)Thought:\s*(.+?)(?=\n|$)`)
	actionPattern := regexp.MustCompile(`(?is)Action:\s*(.+?)(?=\n|$)`)

	if thoughtMatch := thoughtPattern.FindStringSubmatch(text); len(thoughtMatch) > 1 {
		thought = strings.TrimSpace(thoughtMatch[1])
	}
	if actionMatch := actionPattern.FindStringSubmatch(text); len(actionMatch) > 1 {
		action = strings.TrimSpace(actionMatch[1])
	}

	return thought, action
}

func (a *ReActAgent) parseAction(actionText string) (toolName, toolInput string) {
	pattern := regexp.MustCompile(`(\w+)\[(.*)\]`)
	if match := pattern.FindStringSubmatch(actionText); len(match) > 2 {
		return match[1], match[2]
	}
	return "", ""
}

func (a *ReActAgent) parseActionInput(actionText string) string {
	pattern := regexp.MustCompile(`\w+\[(.*)\]`)
	if match := pattern.FindStringSubmatch(actionText); len(match) > 1 {
		return match[1]
	}
	return ""
}

func (a *ReActAgent) Run(ctx context.Context, input string) (string, error) {
	a.CurrentHistory = make([]string, 0)
	currentStep := 0

	fmt.Printf("\n🤖 %s starting: %s\n", a.Name(), input)

	for currentStep < a.MaxSteps {
		currentStep++
		fmt.Printf("\n--- Step %d ---\n", currentStep)

		prompt := a.buildPrompt(input)
		messages := []core.ChatMessage{{Role: "user", Content: prompt}}

		response, err := a.LLM.Invoke(ctx, messages)
		if err != nil {
			fmt.Printf("❌ Error: LLM failed to return valid response.\n")
			break
		}

		thought, action := a.parseOutput(response)

		if thought != "" {
			fmt.Printf("🤔 Thought: %s\n", thought)
		}

		if action == "" {
			fmt.Printf("⚠️ Warning: No valid Action parsed, terminating.\n")
			break
		}

		if strings.HasPrefix(action, "Finish") {
			finalAnswer := a.parseActionInput(action)
			fmt.Printf("🎉 Final Answer: %s\n", finalAnswer)

			a.AddMessage(core.NewUserMessage(input))
			a.AddMessage(core.NewAssistantMessage(finalAnswer))

			return finalAnswer, nil
		}

		toolName, toolInput := a.parseAction(action)
		if toolName == "" || toolInput == "" {
			a.CurrentHistory = append(a.CurrentHistory, "Observation: Invalid Action format, please check.")
			continue
		}

		fmt.Printf("🎬 Action: %s[%s]\n", toolName, toolInput)

		observation := ""
		if a.ToolRegistry != nil {
			observation = a.ToolRegistry.Execute(toolName, map[string]interface{}{"input": toolInput})
		} else {
			observation = "❌ No tool registry configured"
		}
		fmt.Printf("👀 Observation: %s\n", observation)

		a.CurrentHistory = append(a.CurrentHistory, fmt.Sprintf("Action: %s", action))
		a.CurrentHistory = append(a.CurrentHistory, fmt.Sprintf("Observation: %s", observation))
	}

	fmt.Println("⏰ Max steps reached, terminating.")
	finalAnswer := "Sorry, I could not complete this task within the limit."

	a.AddMessage(core.NewUserMessage(input))
	a.AddMessage(core.NewAssistantMessage(finalAnswer))

	return finalAnswer, nil
}

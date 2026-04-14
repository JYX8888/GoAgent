package agents

import (
	"context"
	"fmt"
	"strings"

	"GoAgent/pkg/core"
)

const DefaultInitialPrompt = `Please complete the following task:

Task: {task}

Please provide a complete and accurate answer.`

const DefaultReflectPrompt = `Please carefully review the following answer and identify possible issues or areas for improvement:

Original Task:
{task}

Current Answer:
{content}

Please analyze the quality of this answer, point out shortcomings, and suggest specific improvements.
If the answer is already good, please respond with "No improvements needed".`

const DefaultRefinePrompt = `Please improve your answer based on the feedback:

Original Task:
{task}

Previous Attempt:
{last_attempt}

Feedback:
{feedback}

Please provide an improved answer.`

type Memory struct {
	Records []struct {
		Type    string
		Content string
	}
}

func (m *Memory) AddRecord(recordType, content string) {
	m.Records = append(m.Records, struct {
		Type    string
		Content string
	}{Type: recordType, Content: content})
	fmt.Printf("📝 Memory updated, added '%s' record.\n", recordType)
}

func (m *Memory) GetTrajectory() string {
	var sb strings.Builder
	for _, record := range m.Records {
		if record.Type == "execution" {
			sb.WriteString("--- Previous Attempt ---\n")
			sb.WriteString(record.Content)
			sb.WriteString("\n\n")
		} else if record.Type == "reflection" {
			sb.WriteString("--- Reviewer Feedback ---\n")
			sb.WriteString(record.Content)
			sb.WriteString("\n\n")
		}
	}
	return strings.TrimSpace(sb.String())
}

func (m *Memory) GetLastExecution() string {
	for i := len(m.Records) - 1; i >= 0; i-- {
		if m.Records[i].Type == "execution" {
			return m.Records[i].Content
		}
	}
	return ""
}

type ReflectionAgent struct {
	*core.BaseAgent
	MaxIterations int
	Memory        *Memory
	Prompts       map[string]string
}

type ReflectionAgentOption func(*ReflectionAgent)

func WithMaxIterations(max int) ReflectionAgentOption {
	return func(a *ReflectionAgent) { a.MaxIterations = max }
}

func WithReflectionPrompts(prompts map[string]string) ReflectionAgentOption {
	return func(a *ReflectionAgent) { a.Prompts = prompts }
}

func NewReflectionAgent(name string, llm *core.LLM, systemPrompt string, opts ...ReflectionAgentOption) *ReflectionAgent {
	prompts := map[string]string{
		"initial": DefaultInitialPrompt,
		"reflect": DefaultReflectPrompt,
		"refine":  DefaultRefinePrompt,
	}

	agent := &ReflectionAgent{
		BaseAgent:     core.NewBaseAgent(name, llm, systemPrompt, nil),
		MaxIterations: 3,
		Memory:        &Memory{},
		Prompts:       prompts,
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func (a *ReflectionAgent) getLLMResponse(ctx context.Context, prompt string) (string, error) {
	messages := []core.ChatMessage{{Role: "user", Content: prompt}}
	return a.LLM.Invoke(ctx, messages)
}

func (a *ReflectionAgent) Run(ctx context.Context, input string) (string, error) {
	fmt.Printf("\n🤖 %s starting task: %s\n", a.Name(), input)

	a.Memory = &Memory{}

	fmt.Println("\n--- Initial attempt ---")
	initialPrompt := strings.ReplaceAll(a.Prompts["initial"], "{task}", input)
	initialResult, err := a.getLLMResponse(ctx, initialPrompt)
	if err != nil {
		return "", fmt.Errorf("initial LLM call failed: %w", err)
	}
	a.Memory.AddRecord("execution", initialResult)

	for i := 0; i < a.MaxIterations; i++ {
		fmt.Printf("\n--- Iteration %d/%d ---\n", i+1, a.MaxIterations)

		fmt.Println("\n-> Reflecting...")
		lastResult := a.Memory.GetLastExecution()
		reflectPrompt := a.Prompts["reflect"]
		reflectPrompt = strings.ReplaceAll(reflectPrompt, "{task}", input)
		reflectPrompt = strings.ReplaceAll(reflectPrompt, "{content}", lastResult)
		feedback, err := a.getLLMResponse(ctx, reflectPrompt)
		if err != nil {
			return "", fmt.Errorf("reflect LLM call failed: %w", err)
		}
		a.Memory.AddRecord("reflection", feedback)

		if strings.Contains(strings.ToLower(feedback), "no improvements needed") ||
			strings.Contains(strings.ToLower(feedback), "无需改进") {
			fmt.Println("\n✅ Reflection found no improvements needed, task completed.")
			break
		}

		fmt.Println("\n-> Refining...")
		refinePrompt := a.Prompts["refine"]
		refinePrompt = strings.ReplaceAll(refinePrompt, "{task}", input)
		refinePrompt = strings.ReplaceAll(refinePrompt, "{last_attempt}", lastResult)
		refinePrompt = strings.ReplaceAll(refinePrompt, "{feedback}", feedback)
		refinedResult, err := a.getLLMResponse(ctx, refinePrompt)
		if err != nil {
			return "", fmt.Errorf("refine LLM call failed: %w", err)
		}
		a.Memory.AddRecord("execution", refinedResult)
	}

	finalResult := a.Memory.GetLastExecution()
	fmt.Printf("\n--- Task completed ---\nFinal Result:\n%s\n", finalResult)

	a.AddMessage(core.NewUserMessage(input))
	a.AddMessage(core.NewAssistantMessage(finalResult))

	return finalResult, nil
}

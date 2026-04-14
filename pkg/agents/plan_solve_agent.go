package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"GoAgent/pkg/core"
)

const DefaultPlannerPrompt = `You are a top AI planning expert. Your task is to break down complex user problems into a plan with multiple simple steps.
Make sure each step in the plan is an independent, executable subtask, and strictly arrange them in logical order.
Your output must be a Python list, where each element is a string describing a subtask.

Question: {question}

Please output your plan strictly in the following format:
"""python
["step 1", "step 2", "step 3", ...]
"""`

const DefaultExecutorPrompt = `You are a top AI execution expert. Your task is to strictly follow the given plan and solve problems step by step.
You will receive the original question, the complete plan, and the steps completed so far with their results.
Please focus on solving the "current step" and only output the final answer for that step, without any additional explanation or dialogue.

Original Question:
{question}

Complete Plan:
{plan}

History Steps and Results:
{history}

Current Step:
{current_step}

Please only output the answer for the "current step":`

type Planner struct {
	LLM            *core.LLM
	PromptTemplate string
}

type Executor struct {
	LLM            *core.LLM
	PromptTemplate string
}

func NewPlanner(llm *core.LLM, promptTemplate string) *Planner {
	tpl := DefaultPlannerPrompt
	if promptTemplate != "" {
		tpl = promptTemplate
	}
	return &Planner{LLM: llm, PromptTemplate: tpl}
}

func (p *Planner) Plan(ctx context.Context, question string) ([]string, error) {
	prompt := strings.ReplaceAll(p.PromptTemplate, "{question}", question)
	messages := []core.ChatMessage{{Role: "user", Content: prompt}}

	fmt.Println("--- Generating plan ---")
	response, err := p.LLM.Invoke(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM invoke failed: %w", err)
	}
	fmt.Printf("✅ Plan generated:\n%s\n", response)

	plan := p.parsePlan(response)
	return plan, nil
}

func (p *Planner) parsePlan(response string) []string {
	pattern := regexp.MustCompile(`"""python\s*\[(.*?)\]\s*"""`)
	matches := pattern.FindStringSubmatch(response)

	if len(matches) < 2 {
		pattern2 := regexp.MustCompile(`\[\s*"([^"]+)"`)
		matches2 := pattern2.FindAllStringSubmatch(response, -1)
		if len(matches2) > 0 {
			var steps []string
			for _, m := range matches2 {
				if len(m) > 1 {
					steps = append(steps, m[1])
				}
			}
			return steps
		}
		return nil
	}

	content := matches[1]
	pattern3 := regexp.MustCompile(`"([^"]+)"`)
	steps := pattern3.FindAllStringSubmatch(content, -1)
	var result []string
	for _, s := range steps {
		if len(s) > 1 {
			result = append(result, s[1])
		}
	}
	return result
}

func NewExecutor(llm *core.LLM, promptTemplate string) *Executor {
	tpl := DefaultExecutorPrompt
	if promptTemplate != "" {
		tpl = promptTemplate
	}
	return &Executor{LLM: llm, PromptTemplate: tpl}
}

func (e *Executor) Execute(ctx context.Context, question string, plan []string) (string, error) {
	history := ""
	finalAnswer := ""

	fmt.Println("\n--- Executing plan ---")
	for i, step := range plan {
		fmt.Printf("\n-> Executing step %d/%d: %s\n", i+1, len(plan), step)

		prompt := e.PromptTemplate
		prompt = strings.ReplaceAll(prompt, "{question}", question)
		prompt = strings.ReplaceAll(prompt, "{plan}", strings.Join(plan, ", "))
		hist := history
		if hist == "" {
			hist = "None"
		}
		prompt = strings.ReplaceAll(prompt, "{history}", hist)
		prompt = strings.ReplaceAll(prompt, "{current_step}", step)

		messages := []core.ChatMessage{{Role: "user", Content: prompt}}
		response, err := e.LLM.Invoke(ctx, messages)
		if err != nil {
			return "", fmt.Errorf("LLM invoke failed at step %d: %w", i+1, err)
		}

		history += fmt.Sprintf("Step %d: %s\nResult: %s\n\n", i+1, step, response)
		finalAnswer = response
		fmt.Printf("✅ Step %d completed, result: %s\n", i+1, finalAnswer)
	}

	return finalAnswer, nil
}

type PlanAndSolveAgent struct {
	*core.BaseAgent
	Planner  *Planner
	Executor *Executor
}

type PlanAndSolveOption func(*PlanAndSolveAgent)

func WithPlannerPrompt(prompt string) PlanAndSolveOption {
	return func(a *PlanAndSolveAgent) { a.Planner.PromptTemplate = prompt }
}

func WithExecutorPrompt(prompt string) PlanAndSolveOption {
	return func(a *PlanAndSolveAgent) { a.Executor.PromptTemplate = prompt }
}

func NewPlanAndSolveAgent(name string, llm *core.LLM, systemPrompt string, opts ...PlanAndSolveOption) *PlanAndSolveAgent {
	agent := &PlanAndSolveAgent{
		BaseAgent: core.NewBaseAgent(name, llm, systemPrompt, nil),
		Planner:   NewPlanner(llm, ""),
		Executor:  NewExecutor(llm, ""),
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func (a *PlanAndSolveAgent) Run(ctx context.Context, input string) (string, error) {
	fmt.Printf("\n🤖 %s starting: %s\n", a.Name(), input)

	plan, err := a.Planner.Plan(ctx, input)
	if err != nil || len(plan) == 0 {
		finalAnswer := "Could not generate a valid plan, task terminated."
		fmt.Printf("\n--- Task terminated ---\n%s\n", finalAnswer)

		a.AddMessage(core.NewUserMessage(input))
		a.AddMessage(core.NewAssistantMessage(finalAnswer))

		return finalAnswer, nil
	}

	finalAnswer, err := a.Executor.Execute(ctx, input, plan)
	if err != nil {
		finalAnswer = fmt.Sprintf("Error during execution: %v", err)
	}

	fmt.Printf("\n--- Task completed ---\nFinal Answer: %s\n", finalAnswer)

	a.AddMessage(core.NewUserMessage(input))
	a.AddMessage(core.NewAssistantMessage(finalAnswer))

	return finalAnswer, nil
}

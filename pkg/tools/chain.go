package tools

import (
	"fmt"
	"strings"
	"sync"
)

type ChainStep struct {
	ToolName      string
	InputTemplate string
	OutputKey     string
}

type ToolChain struct {
	Name        string
	Description string
	Steps       []ChainStep
	mu          sync.RWMutex
}

func NewToolChain(name, description string) *ToolChain {
	return &ToolChain{
		Name:        name,
		Description: description,
		Steps:       make([]ChainStep, 0),
	}
}

func (c *ToolChain) AddStep(toolName, inputTemplate, outputKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if outputKey == "" {
		outputKey = fmt.Sprintf("step_%d_result", len(c.Steps))
	}

	step := ChainStep{
		ToolName:      toolName,
		InputTemplate: inputTemplate,
		OutputKey:     outputKey,
	}
	c.Steps = append(c.Steps, step)
	fmt.Printf("✅ Chain '%s' added step: %s\n", c.Name, toolName)
}

func (c *ToolChain) Execute(registry *ToolRegistry, inputData string, context map[string]interface{}) string {
	c.mu.RLock()
	steps := make([]ChainStep, len(c.Steps))
	copy(steps, c.Steps)
	c.mu.RUnlock()

	if len(steps) == 0 {
		return "❌ Chain is empty, cannot execute"
	}

	fmt.Printf("🚀 Starting chain execution: %s\n", c.Name)

	if context == nil {
		context = make(map[string]interface{})
	}
	context["input"] = inputData

	finalResult := inputData

	for i, step := range steps {
		fmt.Printf("📝 Executing step %d/%d: %s\n", i+1, len(steps), step.ToolName)

		actualInput := step.InputTemplate
		for key, value := range context {
			actualInput = strings.ReplaceAll(actualInput, "{"+key+"}", fmt.Sprintf("%v", value))
		}

		result := registry.Execute(step.ToolName, map[string]interface{}{"input": actualInput})
		context[step.OutputKey] = result
		finalResult = result
		fmt.Printf("✅ Step %d completed\n", i+1)
	}

	fmt.Printf("🎉 Chain '%s' completed\n", c.Name)
	return finalResult
}

func (c *ToolChain) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("ToolChain(name=%s, steps=%d)", c.Name, len(c.Steps))
}

type ToolChainManager struct {
	Registry *ToolRegistry
	Chains   map[string]*ToolChain
	mu       sync.RWMutex
}

func NewToolChainManager(registry *ToolRegistry) *ToolChainManager {
	return &ToolChainManager{
		Registry: registry,
		Chains:   make(map[string]*ToolChain),
	}
}

func (m *ToolChainManager) RegisterChain(chain *ToolChain) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Chains[chain.Name] = chain
	fmt.Printf("✅ Chain '%s' registered\n", chain.Name)
}

func (m *ToolChainManager) ExecuteChain(chainName, inputData string, context map[string]interface{}) string {
	m.mu.RLock()
	chain, exists := m.Chains[chainName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Sprintf("❌ Chain '%s' not found", chainName)
	}

	return chain.Execute(m.Registry, inputData, context)
}

func (m *ToolChainManager) ListChains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.Chains))
	for name := range m.Chains {
		result = append(result, name)
	}
	return result
}

func (m *ToolChainManager) GetChainInfo(chainName string) map[string]interface{} {
	m.mu.RLock()
	chain, exists := m.Chains[chainName]
	m.mu.RUnlock()

	if !exists {
		return nil
	}

	chain.mu.RLock()
	defer chain.mu.RUnlock()

	stepDetails := make([]map[string]interface{}, len(chain.Steps))
	for i, step := range chain.Steps {
		stepDetails[i] = map[string]interface{}{
			"tool_name":      step.ToolName,
			"input_template": step.InputTemplate,
			"output_key":     step.OutputKey,
		}
	}

	return map[string]interface{}{
		"name":         chain.Name,
		"description":  chain.Description,
		"steps":        len(chain.Steps),
		"step_details": stepDetails,
	}
}

func CreateResearchChain() *ToolChain {
	chain := NewToolChain("research_and_calculate", "Search information and perform calculations")

	chain.AddStep("search", "{input}", "search_result")
	chain.AddStep("calculator", "2 + 2", "calc_result")

	return chain
}

func CreateSimpleChain() *ToolChain {
	chain := NewToolChain("simple_demo", "Simple chain demo")

	chain.AddStep("calculator", "{input}", "result")

	return chain
}

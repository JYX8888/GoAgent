package tools

import (
	"fmt"
	"sync"
)

type ToolParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
}

func NewToolParameter(name, paramType, description string, required bool, defaultVal interface{}) ToolParameter {
	return ToolParameter{
		Name:        name,
		Type:        paramType,
		Description: description,
		Required:    required,
		Default:     defaultVal,
	}
}

type Tool interface {
	Name() string
	Description() string
	GetParameters() []ToolParameter
	Run(params map[string]interface{}) (string, error)
}

type BaseTool struct {
	Name_        string
	Description_ string
	Parameters_  []ToolParameter
	mu           sync.RWMutex
}

func (t *BaseTool) Name() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Name_
}

func (t *BaseTool) Description() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Description_
}

func (t *BaseTool) GetParameters() []ToolParameter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]ToolParameter, len(t.Parameters_))
	copy(result, t.Parameters_)
	return result
}

func (t *BaseTool) SetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Name_ = name
}

func (t *BaseTool) SetDescription(desc string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Description_ = desc
}

func (t *BaseTool) AddParameter(param ToolParameter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Parameters_ = append(t.Parameters_, param)
}

func (t *BaseTool) ValidateParameters(params map[string]interface{}) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, p := range t.Parameters_ {
		if p.Required {
			if _, ok := params[p.Name]; !ok {
				return false
			}
		}
	}
	return true
}

func (t *BaseTool) ToMap() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	params := make([]map[string]interface{}, len(t.Parameters_))
	for i, p := range t.Parameters_ {
		params[i] = map[string]interface{}{
			"name":        p.Name,
			"type":        p.Type,
			"description": p.Description,
			"required":    p.Required,
			"default":     p.Default,
		}
	}

	return map[string]interface{}{
		"name":        t.Name_,
		"description": t.Description_,
		"parameters":  params,
	}
}

func (t *BaseTool) String() string {
	return fmt.Sprintf("Tool(name=%s)", t.Name_)
}

type ToolFunc func(params map[string]interface{}) (string, error)

type FuncTool struct {
	BaseTool
	Func ToolFunc
}

func NewFuncTool(name, desc string, fn ToolFunc, params ...ToolParameter) *FuncTool {
	tool := &FuncTool{
		BaseTool: BaseTool{
			Name_:        name,
			Description_: desc,
			Parameters_:  params,
		},
		Func: fn,
	}
	return tool
}

func (t *FuncTool) Run(params map[string]interface{}) (string, error) {
	if t.Func == nil {
		return "", fmt.Errorf("tool function not set")
	}
	return t.Func(params)
}

func (t *FuncTool) String() string {
	return fmt.Sprintf("FuncTool(name=%s)", t.Name())
}

package tools

import (
	"fmt"
	"sync"
)

type Tool interface {
	Name() string
	Description() string
	Run(params map[string]interface{}) (string, error)
}

type BaseTool struct {
	Name_        string
	Description_ string
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

type ToolFunc func(params map[string]interface{}) (string, error)

type FuncTool struct {
	BaseTool
	Func ToolFunc
}

func NewFuncTool(name, desc string, fn ToolFunc) *FuncTool {
	return &FuncTool{
		BaseTool: BaseTool{Name_: name, Description_: desc},
		Func:     fn,
	}
}

func (t *FuncTool) Run(params map[string]interface{}) (string, error) {
	if t.Func == nil {
		return "", fmt.Errorf("tool function not set")
	}
	return t.Func(params)
}

package tools

import (
	"fmt"
	"sync"
)

type ToolRegistry struct {
	tools     map[string]Tool
	functions map[string]struct {
		desc string
		fn   func(params map[string]interface{}) (string, error)
	}
	mu sync.RWMutex
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
		functions: make(map[string]struct {
			desc string
			fn   func(params map[string]interface{}) (string, error)
		}),
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name()]; exists {
		fmt.Printf("⚠️ Warning: Tool '%s' already exists, will be replaced.\n", tool.Name())
	}
	r.tools[tool.Name()] = tool
	fmt.Printf("✅ Tool '%s' registered.\n", tool.Name())
}

func (r *ToolRegistry) RegisterFunc(name, desc string, fn func(params map[string]interface{}) (string, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		fmt.Printf("⚠️ Warning: Function tool '%s' already exists, will be replaced.\n", name)
	}
	r.functions[name] = struct {
		desc string
		fn   func(params map[string]interface{}) (string, error)
	}{desc: desc, fn: fn}
	fmt.Printf("✅ Function tool '%s' registered.\n", name)
}

func (r *ToolRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tools[name]; ok {
		delete(r.tools, name)
		fmt.Printf("🗑️ Tool '%s' unregistered.\n", name)
		return true
	}
	if _, ok := r.functions[name]; ok {
		delete(r.functions, name)
		fmt.Printf("🗑️ Function tool '%s' unregistered.\n", name)
		return true
	}
	fmt.Printf("⚠️ Tool '%s' not found.\n", name)
	return false
}

func (r *ToolRegistry) Get(name string) Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

func (r *ToolRegistry) GetFunc(name string) (string, func(params map[string]interface{}) (string, error)) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if fn, ok := r.functions[name]; ok {
		return fn.desc, fn.fn
	}
	return "", nil
}

func (r *ToolRegistry) Execute(name string, params map[string]interface{}) string {
	r.mu.RLock()
	tool, toolExists := r.tools[name]
	fn, fnExists := r.functions[name]
	r.mu.RUnlock()

	if toolExists {
		result, err := tool.Run(params)
		if err != nil {
			return fmt.Sprintf("❌ Error executing tool '%s': %v", name, err)
		}
		return result
	}

	if fnExists {
		result, err := fn.fn(params)
		if err != nil {
			return fmt.Sprintf("❌ Error executing function tool '%s': %v", name, err)
		}
		return result
	}

	return fmt.Sprintf("❌ Tool '%s' not found", name)
}

func (r *ToolRegistry) GetDescription() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var descs []string
	for _, tool := range r.tools {
		descs = append(descs, fmt.Sprintf("- %s: %s", tool.Name(), tool.Description()))
	}
	for name, fn := range r.functions {
		descs = append(descs, fmt.Sprintf("- %s: %s", name, fn.desc))
	}

	if len(descs) == 0 {
		return "No tools available"
	}
	return joinStrings(descs, "\n")
}

func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.tools)+len(r.functions))
	for name := range r.tools {
		result = append(result, name)
	}
	for name := range r.functions {
		result = append(result, name)
	}
	return result
}

func (r *ToolRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]Tool)
	r.functions = make(map[string]struct {
		desc string
		fn   func(params map[string]interface{}) (string, error)
	})
	fmt.Println("🧹 All tools cleared.")
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

var globalRegistry = NewToolRegistry()

func GlobalRegistry() *ToolRegistry {
	return globalRegistry
}

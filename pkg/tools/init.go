package tools

import "fmt"

func RegisterBuiltinTools() {
	registry := GlobalRegistry()

	calculator := NewCalculatorTool()
	registry.Register(calculator)

	search := NewSearchTool()
	registry.Register(search)

	memory := NewMemoryTool()
	registry.Register(memory)

	rag := NewRAGTool()
	registry.Register(rag)

	fmt.Println("✅ Built-in tools registered")
}

func init() {
	RegisterBuiltinTools()
}

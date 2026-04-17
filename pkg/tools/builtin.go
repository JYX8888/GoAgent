package tools

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type CalculatorTool struct {
	BaseTool
}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		BaseTool: BaseTool{
			Name_:        "calculator",
			Description_: "A calculator tool that performs basic arithmetic operations",
			Parameters_: []ToolParameter{
				NewToolParameter("input", "string", "Mathematical expression to calculate", true, nil),
			},
		},
	}
}

func (t *CalculatorTool) Run(params map[string]interface{}) (string, error) {
	input, ok := params["input"].(string)
	if !ok {
		if inputVal, ok := params["input"].(string); !ok {
			return "", fmt.Errorf("invalid input parameter")
		} else {
			input = inputVal
		}
	}

	result := Calculate(input)
	return fmt.Sprintf("Result: %s", result), nil
}

func Calculate(expr string) string {
	expr = strings.TrimSpace(expr)
	expr = strings.ReplaceAll(expr, " ", "")

	result := evaluate(expr)
	if result == math.MaxFloat64 {
		return "Error: Invalid expression"
	}
	return fmt.Sprintf("%v", result)
}

func evaluate(expr string) float64 {
	expr = strings.TrimSpace(expr)

	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		var sum float64
		for _, p := range parts {
			sum += evaluate(p)
		}
		return sum
	}

	if strings.Contains(expr, "-") && !strings.HasPrefix(expr, "-") {
		parts := strings.Split(expr, "-")
		result := evaluate(parts[0])
		for i := 1; i < len(parts); i++ {
			result -= evaluate(parts[i])
		}
		return result
	}

	if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		result := 1.0
		for _, p := range parts {
			result *= evaluate(p)
		}
		return result
	}

	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		result := evaluate(parts[0])
		for i := 1; i < len(parts); i++ {
			divisor := evaluate(parts[i])
			if divisor == 0 {
				return math.MaxFloat64
			}
			result /= divisor
		}
		return result
	}

	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val := evaluate(inner)
		return math.Sqrt(val)
	}

	if strings.HasPrefix(expr, "pow(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 2 {
			base := evaluate(parts[0])
			exp := evaluate(parts[1])
			return math.Pow(base, exp)
		}
	}

	if val, err := strconv.ParseFloat(expr, 64); err == nil {
		return val
	}

	return math.MaxFloat64
}

type SearchTool struct {
	BaseTool
}

func NewSearchTool() *SearchTool {
	return &SearchTool{
		BaseTool: BaseTool{
			Name_:        "search",
			Description_: "A search tool for querying information",
			Parameters_: []ToolParameter{
				NewToolParameter("query", "string", "Search query", true, nil),
			},
		},
	}
}

func (t *SearchTool) Run(params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok {
		if q, ok := params["query"].(string); !ok {
			return "", fmt.Errorf("invalid query parameter")
		} else {
			query = q
		}
	}

	return fmt.Sprintf("Search results for '%s': This is a placeholder search result. In a real implementation, this would connect to a search API.", query), nil
}

type MemoryTool struct {
	BaseTool
}

func NewMemoryTool() *MemoryTool {
	return &MemoryTool{
		BaseTool: BaseTool{
			Name_:        "memory",
			Description_: "A memory tool for storing and retrieving information",
			Parameters_: []ToolParameter{
				NewToolParameter("action", "string", "Action: store, search, recall", true, nil),
				NewToolParameter("content", "string", "Content to store or query", false, nil),
			},
		},
	}
}

func (t *MemoryTool) Run(params map[string]interface{}) (string, error) {
	action, ok := params["action"].(string)
	if !ok {
		return "", fmt.Errorf("invalid action parameter")
	}

	switch action {
	case "store":
		content, _ := params["content"].(string)
		return fmt.Sprintf("Stored: %s", content), nil
	case "search", "recall":
		content, _ := params["content"].(string)
		return fmt.Sprintf("Retrieved memory for: %s", content), nil
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

type RAGTool struct {
	BaseTool
}

func NewRAGTool() *RAGTool {
	return &RAGTool{
		BaseTool: BaseTool{
			Name_:        "rag",
			Description_: "RAG tool for retrieving relevant documents",
			Parameters_: []ToolParameter{
				NewToolParameter("query", "string", "Query to search in documents", true, nil),
				NewToolParameter("action", "string", "Action: search, add_text", false, "search"),
			},
		},
	}
}

func (t *RAGTool) Run(params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok {
		return "", fmt.Errorf("invalid query parameter")
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "search"
	}

	if action == "add_text" {
		return "Document added successfully", nil
	}

	return fmt.Sprintf("RAG search result for '%s': Relevant context would be retrieved here.", query), nil
}

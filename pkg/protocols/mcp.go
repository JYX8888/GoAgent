package protocols

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type MCPContext struct {
	Messages  []map[string]interface{} `json:"messages"`
	Tools     []map[string]interface{} `json:"tools"`
	Resources []map[string]interface{} `json:"resources"`
	Metadata  map[string]interface{}   `json:"metadata"`
}

func NewMCPContext() *MCPContext {
	return &MCPContext{
		Messages:  make([]map[string]interface{}, 0),
		Tools:     make([]map[string]interface{}, 0),
		Resources: make([]map[string]interface{}, 0),
		Metadata:  make(map[string]interface{}),
	}
}

func CreateContext(messages, tools, resources []map[string]interface{}, metadata map[string]interface{}) *MCPContext {
	ctx := NewMCPContext()
	if messages != nil {
		ctx.Messages = messages
	}
	if tools != nil {
		ctx.Tools = tools
	}
	if resources != nil {
		ctx.Resources = resources
	}
	if metadata != nil {
		ctx.Metadata = metadata
	}
	return ctx
}

func ParseContext(data interface{}) (*MCPContext, error) {
	if data == nil {
		return NewMCPContext(), nil
	}

	var ctx MCPContext
	var err error

	switch v := data.(type) {
	case string:
		err = json.Unmarshal([]byte(v), &ctx)
	case []byte:
		err = json.Unmarshal(v, &ctx)
	case map[string]interface{}:
		if bytes, err := json.Marshal(v); err == nil {
			err = json.Unmarshal(bytes, &ctx)
		}
	default:
		return nil, fmt.Errorf("unsupported context type: %T", data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse context: %w", err)
	}

	if ctx.Messages == nil {
		ctx.Messages = make([]map[string]interface{}, 0)
	}
	if ctx.Tools == nil {
		ctx.Tools = make([]map[string]interface{}, 0)
	}
	if ctx.Resources == nil {
		ctx.Resources = make([]map[string]interface{}, 0)
	}
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]interface{})
	}

	return &ctx, nil
}

func (c *MCPContext) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *MCPContext) AddMessage(role, content string) {
	c.Messages = append(c.Messages, map[string]interface{}{
		"role":    role,
		"content": content,
	})
}

func (c *MCPContext) AddTool(name, description string, params map[string]interface{}) {
	c.Tools = append(c.Tools, map[string]interface{}{
		"name":        name,
		"description": description,
		"parameters":  params,
	})
}

func CreateErrorResponse(message, code string, details map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"code":    code,
		},
	}
	if details != nil {
		response["error"].(map[string]interface{})["details"] = details
	}
	return response
}

func CreateSuccessResponse(data interface{}, metadata map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}
	if metadata != nil {
		response["metadata"] = metadata
	}
	return response
}

type MCPTool struct {
	Name        string
	Description string
	Handler     func(params map[string]interface{}) (string, error)
}

type MCPServer struct {
	Name        string
	Description string
	Tools       map[string]*MCPTool
	Resources   map[string]interface{}
	Running     bool
	mu          sync.RWMutex
}

func NewMCPServer(name, description string) *MCPServer {
	return &MCPServer{
		Name:        name,
		Description: description,
		Tools:       make(map[string]*MCPTool),
		Resources:   make(map[string]interface{}),
		Running:     false,
	}
}

func (s *MCPServer) AddTool(name, description string, handler func(params map[string]interface{}) (string, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Tools[name] = &MCPTool{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (s *MCPServer) AddResource(uri string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Resources[uri] = data
}

func (s *MCPServer) GetTools() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(s.Tools))
	for _, tool := range s.Tools {
		result = append(result, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
		})
	}
	return result
}

func (s *MCPServer) GetResources() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(s.Resources))
	for uri, data := range s.Resources {
		result = append(result, map[string]interface{}{
			"uri":  uri,
			"data": data,
		})
	}
	return result
}

func (s *MCPServer) ExecuteTool(name string, params map[string]interface{}) (string, error) {
	s.mu.RLock()
	tool, exists := s.Tools[name]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("tool '%s' not found", name)
	}

	return tool.Handler(params)
}

func (s *MCPServer) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        s.Name,
		"description": s.Description,
		"tools":       s.GetTools(),
		"resources":   s.GetResources(),
	}
}

func (s *MCPServer) Start(addr string) error {
	s.mu.Lock()
	s.Running = true
	s.mu.Unlock()

	fmt.Printf("MCP Server '%s' started at %s\n", s.Name, addr)
	return nil
}

func (s *MCPServer) Stop() error {
	s.mu.Lock()
	s.Running = false
	s.mu.Unlock()

	fmt.Printf("MCP Server '%s' stopped\n", s.Name)
	return nil
}

type MCPClient struct {
	ServerURL string
}

func NewMCPClient(serverURL string) *MCPClient {
	return &MCPClient{ServerURL: serverURL}
}

func (c *MCPClient) CallTool(name string, params map[string]interface{}) (string, error) {
	result, ok := params["input"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input parameter")
	}

	return fmt.Sprintf("MCP Client called tool '%s' with input: %s", name, result), nil
}

func (c *MCPClient) GetServerInfo() (map[string]interface{}, error) {
	return map[string]interface{}{
		"name":       "MCP Client",
		"server_url": c.ServerURL,
		"status":     "connected",
		"timestamp":  time.Now(),
	}, nil
}

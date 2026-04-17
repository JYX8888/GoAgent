package protocols

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type MessageType string

const (
	MessageTypeTask     MessageType = "task"
	MessageTypeResult   MessageType = "result"
	MessageTypeQuery    MessageType = "query"
	MessageTypeResponse MessageType = "response"
	MessageTypeError    MessageType = "error"
)

type A2AMessage struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewA2AMessage(msgType MessageType, from, to string, payload map[string]interface{}) *A2AMessage {
	return &A2AMessage{
		ID:        generateID(),
		Type:      msgType,
		From:      from,
		To:        to,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

func CreateMessage(msgType MessageType, from, to string, payload map[string]interface{}) *A2AMessage {
	return NewA2AMessage(msgType, from, to, payload)
}

func ParseMessage(data []byte) (*A2AMessage, error) {
	var msg A2AMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}

func (m *A2AMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

type Skill struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, payload map[string]interface{}) (string, error)
}

type A2AServer struct {
	Name         string
	Description  string
	Version      string
	Capabilities map[string]interface{}
	Skills       map[string]*Skill
	AgentURL     string
	mu           sync.RWMutex
	httpServer   *http.Server
}

func NewA2AServer(name, description, version string, capabilities map[string]interface{}) *A2AServer {
	return &A2AServer{
		Name:         name,
		Description:  description,
		Version:      version,
		Capabilities: capabilities,
		Skills:       make(map[string]*Skill),
	}
}

func (s *A2AServer) AddSkill(name, description string, handler func(ctx context.Context, payload map[string]interface{}) (string, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Skills[name] = &Skill{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (s *A2AServer) GetInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	skills := make([]map[string]interface{}, 0, len(s.Skills))
	for _, skill := range s.Skills {
		skills = append(skills, map[string]interface{}{
			"name":        skill.Name,
			"description": skill.Description,
		})
	}

	return map[string]interface{}{
		"name":         s.Name,
		"description":  s.Description,
		"version":      s.Version,
		"capabilities": s.Capabilities,
		"skills":       skills,
	}
}

func (s *A2AServer) handleTask(w http.ResponseWriter, r *http.Request) {
	var msg A2AMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.executeSkill(msg.Payload)
	if err != nil {
		respMsg := NewA2AMessage(MessageTypeError, s.Name, msg.From, map[string]interface{}{
			"error": err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(respMsg)
		return
	}

	respMsg := NewA2AMessage(MessageTypeResult, s.Name, msg.From, map[string]interface{}{
		"result": result,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMsg)
}

func (s *A2AServer) executeSkill(payload map[string]interface{}) (string, error) {
	skillName, ok := payload["skill"].(string)
	if !ok {
		return "", fmt.Errorf("skill not specified")
	}

	s.mu.RLock()
	skill, exists := s.Skills[skillName]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("skill '%s' not found", skillName)
	}

	ctx := context.Background()
	return skill.Handler(ctx, payload)
}

func (s *A2AServer) Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.GetInfo())
	})

	mux.HandleFunc("/skills", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		skills := s.Skills
		s.mu.RUnlock()

		skillList := make([]map[string]interface{}, 0)
		for _, skill := range skills {
			skillList = append(skillList, map[string]interface{}{
				"name":        skill.Name,
				"description": skill.Description,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skillList)
	})

	mux.HandleFunc("/tasks", s.handleTask)

	s.httpServer = &http.Server{Addr: addr, Handler: mux}
	return s.httpServer.ListenAndServe()
}

func (s *A2AServer) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

type A2AClient struct {
	AgentURL string
	Client   *http.Client
	mu       sync.RWMutex
}

func NewA2AClient(agentURL string) *A2AClient {
	return &A2AClient{
		AgentURL: agentURL,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *A2AClient) SendTask(ctx context.Context, skillName string, payload map[string]interface{}) (*A2AMessage, error) {
	c.mu.RLock()
	agentURL := c.AgentURL
	c.mu.RUnlock()

	payload["skill"] = skillName

	msg := NewA2AMessage(MessageTypeTask, "client", "agent", payload)

	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", agentURL+"/tasks", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := c.Client
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var resultMsg A2AMessage
	if err := json.NewDecoder(resp.Body).Decode(&resultMsg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resultMsg, nil
}

func (c *A2AClient) GetAgentInfo() (map[string]interface{}, error) {
	c.mu.RLock()
	agentURL := c.AgentURL
	c.mu.RUnlock()

	resp, err := c.Client.Get(agentURL + "/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return info, nil
}

package protocols

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AgentInfo struct {
	Name         string
	Description  string
	Version      string
	URL          string
	Skills       []string
	Capabilities map[string]interface{}
	RegisteredAt time.Time
	LastSeen     time.Time
}

type AgentRegistry struct {
	Agents map[string]*AgentInfo
	mu     sync.RWMutex
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		Agents: make(map[string]*AgentInfo),
	}
}

func (r *AgentRegistry) Register(agent *AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	r.Agents[agent.Name] = agent
}

func (r *AgentRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Agents[name]; exists {
		delete(r.Agents, name)
		return true
	}
	return false
}

func (r *AgentRegistry) Get(name string) *AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Agents[name]
}

func (r *AgentRegistry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*AgentInfo, 0, len(r.Agents))
	for _, agent := range r.Agents {
		result = append(result, agent)
	}
	return result
}

func (r *AgentRegistry) FindBySkill(skillName string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*AgentInfo
	for _, agent := range r.Agents {
		for _, skill := range agent.Skills {
			if skill == skillName {
				result = append(result, agent)
				break
			}
		}
	}
	return result
}

func (r *AgentRegistry) UpdateLastSeen(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent, exists := r.Agents[name]; exists {
		agent.LastSeen = time.Now()
	}
}

type AgentNetwork struct {
	Registry   *AgentRegistry
	ClientPool map[string]*A2AClient
	mu         sync.RWMutex
}

func NewAgentNetwork() *AgentNetwork {
	return &AgentNetwork{
		Registry:   NewAgentRegistry(),
		ClientPool: make(map[string]*A2AClient),
	}
}

func (n *AgentNetwork) AddAgent(agent *AgentInfo) {
	n.Registry.Register(agent)

	n.mu.Lock()
	if _, exists := n.ClientPool[agent.Name]; !exists {
		n.ClientPool[agent.Name] = NewA2AClient(agent.URL)
	}
	n.mu.Unlock()
}

func (n *AgentNetwork) RemoveAgent(name string) bool {
	n.Registry.Unregister(name)

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.ClientPool[name]; exists {
		delete(n.ClientPool, name)
		return true
	}
	return false
}

func (n *AgentNetwork) SendToAgent(ctx context.Context, agentName, skillName string, payload map[string]interface{}) (string, error) {
	n.mu.RLock()
	client, exists := n.ClientPool[agentName]
	n.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("agent '%s' not found in network", agentName)
	}

	msg, err := client.SendTask(ctx, skillName, payload)
	if err != nil {
		return "", err
	}

	if msg.Type == MessageTypeError {
		if errPayload, ok := msg.Payload["error"].(string); ok {
			return "", fmt.Errorf("agent error: %s", errPayload)
		}
		return "", fmt.Errorf("unknown agent error")
	}

	if result, ok := msg.Payload["result"].(string); ok {
		return result, nil
	}

	return "", fmt.Errorf("invalid response format")
}

func (n *AgentNetwork) Broadcast(ctx context.Context, skillName string, payload map[string]interface{}) map[string]string {
	results := make(map[string]string)

	agents := n.Registry.List()
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, agent := range agents {
		wg.Add(1)
		go func(a *AgentInfo) {
			defer wg.Done()

			result, err := n.SendToAgent(ctx, a.Name, skillName, payload)
			mu.Lock()
			if err != nil {
				results[a.Name] = fmt.Sprintf("Error: %v", err)
			} else {
				results[a.Name] = result
			}
			mu.Unlock()
		}(agent)
	}

	wg.Wait()
	return results
}

func (n *AgentNetwork) FindAgentsBySkill(skillName string) []*AgentInfo {
	return n.Registry.FindBySkill(skillName)
}

func (n *AgentNetwork) ListAgents() []*AgentInfo {
	return n.Registry.List()
}

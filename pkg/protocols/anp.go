package protocols

import (
	"fmt"
	"sync"
	"time"
)

type ServiceInfo struct {
	ServiceID     string                 `json:"service_id"`
	ServiceType   string                 `json:"service_type"`
	Endpoint      string                 `json:"endpoint"`
	ServiceName   string                 `json:"service_name"`
	Capabilities  []string               `json:"capabilities"`
	Metadata      map[string]interface{} `json:"metadata"`
	RegisteredAt  time.Time              `json:"registered_at"`
	LastHeartbeat time.Time              `json:"last_heartbeat"`
}

func NewServiceInfo(serviceID, serviceType, endpoint, serviceName string, capabilities []string, metadata map[string]interface{}) *ServiceInfo {
	return &ServiceInfo{
		ServiceID:     serviceID,
		ServiceType:   serviceType,
		Endpoint:      endpoint,
		ServiceName:   serviceName,
		Capabilities:  capabilities,
		Metadata:      metadata,
		RegisteredAt:  time.Now(),
		LastHeartbeat: time.Now(),
	}
}

func (s *ServiceInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"service_id":     s.ServiceID,
		"service_type":   s.ServiceType,
		"endpoint":       s.Endpoint,
		"service_name":   s.ServiceName,
		"capabilities":   s.Capabilities,
		"metadata":       s.Metadata,
		"registered_at":  s.RegisteredAt,
		"last_heartbeat": s.LastHeartbeat,
	}
}

func (s *ServiceInfo) UpdateHeartbeat() {
	s.LastHeartbeat = time.Now()
}

type ANPDiscovery struct {
	Services map[string]*ServiceInfo
	mu       sync.RWMutex
}

func NewANPDiscovery() *ANPDiscovery {
	return &ANPDiscovery{
		Services: make(map[string]*ServiceInfo),
	}
}

func (d *ANPDiscovery) RegisterService(service *ServiceInfo) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	service.RegisteredAt = time.Now()
	service.LastHeartbeat = time.Now()
	d.Services[service.ServiceID] = service
	return true
}

func (d *ANPDiscovery) UnregisterService(serviceID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Services[serviceID]; exists {
		delete(d.Services, serviceID)
		return true
	}
	return false
}

func (d *ANPDiscovery) DiscoverServices(serviceType string, filters map[string]interface{}) []*ServiceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result []*ServiceInfo
	for _, service := range d.Services {
		if serviceType != "" && service.ServiceType != serviceType {
			continue
		}
		result = append(result, service)
	}
	return result
}

func (d *ANPDiscovery) GetService(serviceID string) *ServiceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Services[serviceID]
}

func (d *ANPDiscovery) ListAllServices() []*ServiceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*ServiceInfo, 0, len(d.Services))
	for _, service := range d.Services {
		result = append(result, service)
	}
	return result
}

func (d *ANPDiscovery) Heartbeat(serviceID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if service, exists := d.Services[serviceID]; exists {
		service.LastHeartbeat = time.Now()
		return true
	}
	return false
}

func RegisterService(discovery *ANPDiscovery, service *ServiceInfo) bool {
	return discovery.RegisterService(service)
}

func DiscoverService(discovery *ANPDiscovery, serviceType string) []*ServiceInfo {
	return discovery.DiscoverServices(serviceType, nil)
}

type ANPNetwork struct {
	Discovery *ANPDiscovery
	Services  map[string]*ServiceClient
	mu        sync.RWMutex
}

func NewANPNetwork() *ANPNetwork {
	return &ANPNetwork{
		Discovery: NewANPDiscovery(),
		Services:  make(map[string]*ServiceClient),
	}
}

func (n *ANPNetwork) AddService(service *ServiceInfo) {
	n.Discovery.RegisterService(service)

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.Services[service.ServiceID]; !exists {
		n.Services[service.ServiceID] = NewServiceClient(service.Endpoint)
	}
}

func (n *ANPNetwork) RemoveService(serviceID string) bool {
	n.Discovery.UnregisterService(serviceID)

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.Services[serviceID]; exists {
		delete(n.Services, serviceID)
		return true
	}
	return false
}

func (n *ANPNetwork) GetService(serviceID string) *ServiceInfo {
	return n.Discovery.GetService(serviceID)
}

func (n *ANPNetwork) FindByType(serviceType string) []*ServiceInfo {
	return n.Discovery.DiscoverServices(serviceType, nil)
}

func (n *ANPNetwork) ListServices() []*ServiceInfo {
	return n.Discovery.ListAllServices()
}

type ServiceClient struct {
	Endpoint string
}

func NewServiceClient(endpoint string) *ServiceClient {
	return &ServiceClient{Endpoint: endpoint}
}

func (c *ServiceClient) Call(method string, payload map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Service call to %s, method: %s", c.Endpoint, method),
	}, nil
}

func (c *ServiceClient) Close() error {
	return nil
}

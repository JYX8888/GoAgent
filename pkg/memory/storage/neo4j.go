package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Neo4jClient struct {
	driver neo4j.Driver
	mu     sync.Mutex
}

type Neo4jStorageConfig struct {
	URI                    string
	Username               string
	Password               string
	Database               string
	MaxConnLifetime        int
	MaxConnPoolSize        int
	ConnAcquisitionTimeout int
}

func LoadNeo4jStorageConfigFromEnv() *Neo4jStorageConfig {
	return &Neo4jStorageConfig{
		URI:                    getMySQLEnvOrDefault("NEO4J_URI", "bolt://localhost:7687"),
		Username:               getMySQLEnvOrDefault("NEO4J_USERNAME", "neo4j"),
		Password:               getMySQLEnvOrDefault("NEO4J_PASSWORD", "neo4j"),
		Database:               getMySQLEnvOrDefault("NEO4J_DATABASE", "neo4j"),
		MaxConnLifetime:        3600,
		MaxConnPoolSize:        50,
		ConnAcquisitionTimeout: 60,
	}
}

func NewNeo4jClient(cfg *Neo4jStorageConfig) (*Neo4jClient, error) {
	auth := neo4j.BasicAuth(cfg.Username, cfg.Password, "")
	driver, err := neo4j.NewDriver(cfg.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}
	return &Neo4jClient{driver: driver}, nil
}

func (c *Neo4jClient) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.driver.Close(ctx)
}

func (c *Neo4jClient) ExecuteQuery(ctx context.Context, query string, params map[string]interface{}) (neo4j.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	return session.Run(ctx, query, params)
}

func (c *Neo4jClient) CreateNode(ctx context.Context, label string, props map[string]interface{}) error {
	query := fmt.Sprintf("CREATE (n:%s $props)", label)
	_, err := c.ExecuteQuery(ctx, query, map[string]interface{}{"props": props})
	return err
}

func (c *Neo4jClient) CreateRelationship(ctx context.Context, fromID, toID, relType string, props map[string]interface{}) error {
	query := fmt.Sprintf(`
		MATCH (a), (b)
		WHERE a.id = $from AND b.id = $to
		CREATE (a)-[r:%s]->(b)
		SET r = $props
	`, relType)
	_, err := c.ExecuteQuery(ctx, query, map[string]interface{}{
		"from":  fromID,
		"to":    toID,
		"props": props,
	})
	return err
}

func (c *Neo4jClient) FindNode(ctx context.Context, label string, property string, value interface{}) (map[string]interface{}, error) {
	query := fmt.Sprintf("MATCH (n:%s { %s: $value }) RETURN n", label, property)
	result, err := c.ExecuteQuery(ctx, query, map[string]interface{}{"value": value})
	if err != nil {
		return nil, err
	}

	if result.Next(ctx) {
		record := result.Record()
		if val, ok := record.Get("n"); ok {
			nodeMap, ok := val.(map[string]interface{})
			if ok {
				return nodeMap, nil
			}
		}
	}
	return nil, nil
}

func (c *Neo4jClient) FindNeighbors(ctx context.Context, nodeID string, depth int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("MATCH path = (n)-[*1..%d]-(m) WHERE n.id = $id RETURN m", depth)
	result, err := c.ExecuteQuery(ctx, query, map[string]interface{}{"id": nodeID})
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		if val, ok := record.Get("m"); ok {
			if nodeMap, ok := val.(map[string]interface{}); ok {
				results = append(results, nodeMap)
			}
		}
	}
	return results, nil
}

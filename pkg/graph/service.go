package graph

import (
	"context"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Service struct {
	driver neo4j.DriverWithContext
}

func NewService(ctx context.Context, envSvc *env.Service) (*Service, error) {
	neo4jConfig := envSvc.GetNeo4jConfig()

	driver, err := neo4j.NewDriverWithContext(
		neo4jConfig.URL,
		neo4j.BasicAuth(neo4jConfig.Username, neo4jConfig.Password, "neo4j"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		driver.Close(ctx) // Close only if verification fails
		return nil, fmt.Errorf("failed to verify Neo4j connection: %w", err)
	}
	fmt.Println("Connection established.")

	return &Service{
		driver: driver,
	}, nil
}

// Close closes the Neo4j driver connection
func (s *Service) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

type PersonNode struct {
	OriginalID string `json:"original_id"`
	Username   string `json:"username"`
}

// CreatePerson creates a new person node in the graph
func (s *Service) CreatePerson(ctx context.Context, person PersonNode) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := "CREATE (p:Person {username: $username, original_id: $original_id})"
		params := map[string]any{
			"username":    person.Username,
			"original_id": person.OriginalID,
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// CreateRelationship creates a relationship between two people
func (s *Service) CreateRelationship(ctx context.Context, person1, person2, relationshipType string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (p1:Person {original_id: $person1})
			MATCH (p2:Person {original_id: $person2})
			CREATE (p1)-[r:` + relationshipType + `]->(p2)
		`
		params := map[string]any{
			"person1": person1,
			"person2": person2,
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// GetShortestConnection returns the shortest path between two people by original_id
func (s *Service) GetShortestConnection(ctx context.Context, fromID, toID string) ([]string, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (start:Person {original_id: $fromID}),
				  (end:Person {original_id: $toID}),
				  path = shortestPath((start)-[*]-(end))
			RETURN [node IN nodes(path) | node.username] as path
		`
		params := map[string]any{
			"fromID": fromID,
			"toID":   toID,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var path []string
		if result.Next(ctx) {
			record := result.Record()
			if pathNodes, ok := record.Get("path"); ok {
				path = make([]string, len(pathNodes.([]any)))
				for i, node := range pathNodes.([]any) {
					path[i] = node.(string)
				}
			}
		}
		return path, result.Err()
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return []string{}, nil
	}

	return result.([]string), nil
}

// GetPersonConnections returns all connections for a given person
func (s *Service) GetPersonConnections(ctx context.Context, name string) ([]string, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (p:Person {name: $name})-[r]-(connected:Person)
			RETURN connected.name as name
		`
		params := map[string]any{
			"name": name,
		}
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var connections []string
		for result.Next(ctx) {
			record := result.Record()
			if name, ok := record.Get("name"); ok {
				connections = append(connections, name.(string))
			}
		}
		return connections, result.Err()
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// ClearDatabase removes all nodes and relationships from the database
func (s *Service) ClearDatabase(ctx context.Context) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := "MATCH (n) DETACH DELETE n"
		_, err := tx.Run(ctx, query, nil)
		return nil, err
	})

	return err
}

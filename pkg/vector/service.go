package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

// Point represents a data point to be stored in the vector database
type Point struct {
	ID      string
	Payload map[string]any
	Vector  []float32
}

// PointToUpsert represents a point ready to be upserted to Qdrant
type PointToUpsert struct {
	ID      string
	Payload map[string]interface{}
	Vector  []float32
}

// Service handles vector operations using Qdrant
type Service struct {
	client      *qdrant.Client
	openAISvc   *ai.Service
	collections []string
	dimensions  uint64
}

// NewService creates a new VectorService instance
func NewService(envSvc *env.Service, openAISvc *ai.Service, dimensions uint64) (*Service, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   envSvc.GetQdrantURL(),
		Port:   6334,
		APIKey: envSvc.GetQdrantAPIKey(),
		UseTLS: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	// Test the connection with a simple health check
	ctx := context.Background()
	_, err = client.HealthCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	collections, err := client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return &Service{
		client:      client,
		openAISvc:   openAISvc,
		collections: collections,
		dimensions:  dimensions,
	}, nil
}

// GetCollections returns a list of all collections
func (s *Service) GetCollections() []string {
	return s.collections
}

// EnsureCollection ensures that a collection exists
func (s *Service) EnsureCollectionExists(ctx context.Context, collection string) error {
	exists := false
	for _, name := range s.collections {
		if name == collection {
			exists = true
			break
		}
	}

	if !exists {
		err := s.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: collection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     s.dimensions,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return nil
}

// InitializeCollectionWithData initializes a collection with data if it doesn't exist
func (s *Service) InitializeCollection(ctx context.Context, collection string, points []Point) error {
	if err := s.EnsureCollectionExists(ctx, collection); err != nil {
		return err
	}

	pointsToUpsert := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		pointsToUpsert[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(point.ID),
			Payload: qdrant.NewValueMap(point.Payload),
			Vectors: qdrant.NewVectorsDense(point.Vector),
		}
	}

	upsertPoints := &qdrant.UpsertPoints{
		CollectionName: collection,
		Points:         pointsToUpsert,
	}

	_, err := s.client.Upsert(ctx, upsertPoints)
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}

	return nil
}

// savePointsToFile saves points to a JSON file
func (s *Service) savePointsToFile(points []*qdrant.PointStruct) error {
	pointsFilePath := filepath.Join("data", "points.json")
	if err := os.MkdirAll(filepath.Dir(pointsFilePath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	pointsJSON, err := json.MarshalIndent(points, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling points: %w", err)
	}

	if err := os.WriteFile(pointsFilePath, pointsJSON, 0644); err != nil {
		return fmt.Errorf("error writing points file: %w", err)
	}

	return nil
}

// AddPoints adds points to a collection
func (s *Service) AddPoints(ctx context.Context, collection string, points []struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}) error {
	pointsToUpsert := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		embedding, err := s.openAISvc.CreateJinaEmbedding(point.Text)
		if err != nil {
			return fmt.Errorf("error creating embedding: %w", err)
		}

		pointID := point.ID
		if pointID == "" {
			pointID = uuid.New().String()
		}

		pointsToUpsert[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(pointID),
			Payload: qdrant.NewValueMap(map[string]any{"text": point.Text, "metadata": point.Metadata}),
			Vectors: qdrant.NewVectorsDense(embedding),
		}
	}

	// Save points to file
	if err := s.savePointsToFile(pointsToUpsert); err != nil {
		return fmt.Errorf("error saving points to file: %w", err)
	}

	// Upsert points to Qdrant
	upsertPoints := &qdrant.UpsertPoints{
		CollectionName: collection,
		Points:         pointsToUpsert,
	}

	_, err := s.client.Upsert(ctx, upsertPoints)
	if err != nil {
		return fmt.Errorf("error upserting points: %w", err)
	}

	return nil
}

// PerformSearch performs a vector search
func (s *Service) Search(ctx context.Context, collection string, query string, limit int) ([]Point, error) {
	queryEmbedding, err := s.openAISvc.CreateJinaEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("error creating query embedding: %w", err)
	}

	results, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQueryDense(queryEmbedding),
	})
	if err != nil {
		return nil, fmt.Errorf("error querying points: %w", err)
	}

	points := make([]Point, len(results))
	for i, result := range results {
		payload := make(map[string]interface{})
		for k, v := range result.Payload {
			payload[k] = v.GetStringValue()
		}

		points[i] = Point{
			ID:      result.Id.GetUuid(),
			Payload: payload,
			Vector:  result.Vectors.GetVector().Data,
		}
	}

	return points, nil
}

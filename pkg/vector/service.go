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
func (s *Service) savePointsToFile(points []*qdrant.PointStruct, collection string) error {
	pointsFilePath := filepath.Join("data", "vector", collection+".json")
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

type NewPoint struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}

// AddPoints adds points to a collection and returns the added points
func (s *Service) AddPoints(ctx context.Context, collection string, points []NewPoint) ([]*qdrant.PointStruct, error) {
	fmt.Println("🔍 Adding points to collection:", collection)
	if err := s.EnsureCollectionExists(ctx, collection); err != nil {
		return nil, err
	}
	pointsToUpsert := make([]*qdrant.PointStruct, len(points))
	for i, point := range points {
		fmt.Println("🔍 Creating embedding for point:", point.Metadata)
		embedding, err := s.openAISvc.CreateJinaEmbedding(point.Text)
		if err != nil {
			return nil, fmt.Errorf("error creating embedding: %w", err)
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
	fmt.Println("🔍 Saving points to file")
	if err := s.savePointsToFile(pointsToUpsert, collection); err != nil {
		return nil, fmt.Errorf("error saving points to file: %w", err)
	}

	// Upsert points to Qdrant
	fmt.Println("🔍 Upserting points to Qdrant")
	upsertPoints := &qdrant.UpsertPoints{
		CollectionName: collection,
		Points:         pointsToUpsert,
	}

	_, err := s.client.Upsert(ctx, upsertPoints)
	if err != nil {
		return nil, fmt.Errorf("error upserting points: %w", err)
	}

	return pointsToUpsert, nil
}

// PerformSearch performs a vector search
func (s *Service) Search(ctx context.Context, collection string, query string, limit uint64) ([]*qdrant.ScoredPoint, error) {
	fmt.Println("🔍 Searching for query:", query)
	queryEmbedding, err := s.openAISvc.CreateJinaEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("error creating query embedding: %w", err)
	}

	results, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQueryDense(queryEmbedding),
		Limit:          &limit,
	})

	if err != nil {
		return nil, fmt.Errorf("error querying points: %w", err)
	}

	return results, nil
}

func (s *Service) DeleteCollection(ctx context.Context, collection string) error {
	return s.client.DeleteCollection(ctx, collection)
}

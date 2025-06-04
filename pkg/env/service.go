package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// EnvService handles environment configuration
type Service struct {
	openAIKey     string
	xyzURL        string
	myAPIKey      string
	poligonURL    string
	c3ntralaURL   string
	qdrantURL     string
	qdrantAPIKey  string
	jinaAPIKey    string
	neo4jURL      string
	neo4jUsername string
	neo4jPassword string
	softoURL      string
}

// validateEnv checks if all required environment variables are set
func validateEnv(envVars map[string]string) error {
	var missing []string
	for name, value := range envVars {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// NewEnvService creates a new environment service
func NewService() (*Service, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	envVars := map[string]string{
		"OPENAI_API_KEY": os.Getenv("OPENAI_API_KEY"),
		"XYZ_URL":        os.Getenv("XYZ_URL"),
		"MY_API_KEY":     os.Getenv("MY_API_KEY"),
		"POLIGON_URL":    os.Getenv("POLIGON_URL"),
		"C3NTRALA_URL":   os.Getenv("C3NTRALA_URL"),
		"QDRANT_URL":     os.Getenv("QDRANT_URL"),
		"QDRANT_API_KEY": os.Getenv("QDRANT_API_KEY"),
		"JINA_API_KEY":   os.Getenv("JINA_API_KEY"),
		"NEO4J_URL":      os.Getenv("NEO4J_URL"),
		"NEO4J_USERNAME": os.Getenv("NEO4J_USERNAME"),
		"NEO4J_PASSWORD": os.Getenv("NEO4J_PASSWORD"),
		"SOFTO_URL":      os.Getenv("SOFTO_URL"),
	}

	if err := validateEnv(envVars); err != nil {
		panic(fmt.Sprintf("Environment validation failed: %v", err))
	}

	return &Service{
		openAIKey:     envVars["OPENAI_API_KEY"],
		xyzURL:        envVars["XYZ_URL"],
		myAPIKey:      envVars["MY_API_KEY"],
		poligonURL:    envVars["POLIGON_URL"],
		c3ntralaURL:   envVars["C3NTRALA_URL"],
		qdrantURL:     envVars["QDRANT_URL"],
		qdrantAPIKey:  envVars["QDRANT_API_KEY"],
		jinaAPIKey:    envVars["JINA_API_KEY"],
		neo4jURL:      envVars["NEO4J_URL"],
		neo4jUsername: envVars["NEO4J_USERNAME"],
		neo4jPassword: envVars["NEO4J_PASSWORD"],
		softoURL:      envVars["SOFTO_URL"],
	}, nil
}

// GetOpenAIKey returns the OpenAI API key
func (s *Service) GetOpenAIKey() string {
	return s.openAIKey
}

// GetXYZURL returns the XYZ URL
func (s *Service) GetXYZURL() string {
	return s.xyzURL
}

// GetMyAPIKey returns the API key
func (s *Service) GetMyAPIKey() string {
	return s.myAPIKey
}

// GetPoligonURL returns the Poligon URL
func (s *Service) GetPoligonURL() string {
	return s.poligonURL
}

// GetC3ntralaURL returns the C3ntrala URL
func (s *Service) GetC3ntralaURL() string {
	return s.c3ntralaURL
}

// GetQdrantURL returns the Qdrant URL
func (s *Service) GetQdrantURL() string {
	return s.qdrantURL
}

// GetQdrantAPIKey returns the Qdrant API key
func (s *Service) GetQdrantAPIKey() string {
	return s.qdrantAPIKey
}

// GetJinaAPIKey returns the Jina API key
func (s *Service) GetJinaAPIKey() string {
	return s.jinaAPIKey
}

func (s *Service) GetSoftoURL() string {
	return s.softoURL
}

func (s *Service) GetNeo4jConfig() struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
} {
	return struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		URL:      s.neo4jURL,
		Username: s.neo4jUsername,
		Password: s.neo4jPassword,
	}
}

package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
	return nil
}

func GetOpenAIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func GetRobotGuardURL() string {
	return os.Getenv("ROBOT_GUARD_URL")
}

func GetXYZURL() string {
	return os.Getenv("XYZ_URL")
}

func GetMyAPIKey() string {
	return os.Getenv("MY_API_KEY")
}

func GetPoligonURL() string {
	return os.Getenv("POLIGON_URL")
}

func GetC3ntralaURL() string {
	return os.Getenv("C3NTRALA_URL")
}

func GetQdrantURL() string {
	return os.Getenv("QDRANT_URL")
}

func GetQdrantAPIKey() string {
	return os.Getenv("QDRANT_KEY")
}

func GetJinaAPIKey() string {
	return os.Getenv("JINA_API_KEY")
}

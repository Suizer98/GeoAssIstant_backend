package app

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	UNAMEDB string
	PASSDB  string
	HOSTDB  string
	DBNAME  string
)

func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to default values")
	}

	// Read values from the environment
	UNAMEDB = getEnv("DB_USERNAME", "postgres")
	PASSDB = getEnv("DB_PASSWORD", "postgres123")
	HOSTDB = getEnv("DB_HOST", "postgres")
	DBNAME = getEnv("DB_NAME", "geoaistore")
}

// Helper function to get environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

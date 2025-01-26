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
	DBURL   string
)

func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to default values")
	}

	// Read individual values from the environment
	UNAMEDB = getEnv("DB_USERNAME", "postgres")
	PASSDB = getEnv("DB_PASSWORD", "postgres123")
	HOSTDB = getEnv("DB_HOST", "postgres")
	DBNAME = getEnv("DB_NAME", "geoaistore")

	// Read the full database URL from the environment, if available
	DBURL = getEnv("DB_URL", "")

	// If DB_URL is not provided, construct it from individual components
	if DBURL == "" {
		DBURL = constructDBURL(UNAMEDB, PASSDB, HOSTDB, DBNAME)
	}
}

// constructDBURL constructs the PostgreSQL connection string
func constructDBURL(username, password, host, dbname string) string {
	return "postgres://" + username + ":" + password + "@" + host + "/" + dbname + "?sslmode=disable"
}

// Helper function to get environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

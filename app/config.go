package app

import (
	"log"
	"os"
	"strings"

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
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to default values")
	}

	UNAMEDB = getEnv("UNAMEDB", "postgres")
	PASSDB = getEnv("PASSDB", "postgres123")
	HOSTDB = getEnv("HOSTDB", "postgres")
	DBNAME = getEnv("DBNAME", "geoaistore")
	DBURL = getEnv("DB_URL", "")

	// Incomplete placeholders like "postgresql://" should not win over constructed URL
	if DBURL == "" || !strings.Contains(DBURL, "@") {
		DBURL = constructDBURL(UNAMEDB, PASSDB, HOSTDB, DBNAME)
	}
}

func constructDBURL(username, password, host, dbname string) string {
	return "postgres://" + username + ":" + password + "@" + host + "/" + dbname + "?sslmode=disable"
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

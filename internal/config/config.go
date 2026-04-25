package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           string
	MongoURI       string
	DBName         string
	JWTSecret      string
	JWTExpiryHours int
	UploadDir      string
	GinMode        string
	AdminEmail     string
	AdminPassword  string
}

var AppConfig Config

// Load reads the .env file and populates AppConfig.
func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, falling back to OS environment variables")
	}

	expiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		expiryHours = 24
	}

	AppConfig = Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:         getEnv("DB_NAME", "pocketful_kyc"),
		JWTSecret:      getEnv("JWT_SECRET", "changeme"),
		JWTExpiryHours: expiryHours,
		UploadDir:      getEnv("UPLOAD_DIR", "./uploads"),
		GinMode:        getEnv("GIN_MODE", "debug"),
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@pocketful.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "Admin@12345"),
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(AppConfig.UploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

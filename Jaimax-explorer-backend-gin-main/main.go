package main

import (
	"blockchain-api/config" //These are the local folder of config for configurations 
	"blockchain-api/routes"//These are the local folder of routes for routed use in the code  
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)
// @title           Jaimax Explorer API
// @version         1.0
// @description     Blockchain explorer REST API
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found,please update it's necessary and important for production")
	}

	// Connect to database
	config.ConnectDB()

	// Set Gin mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	// Create router
	r := gin.New()
	r.Use(gin.Recovery())

	// Register all routes
	routes.Register(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Printf("📖 API base path: http://localhost:%s/api/v1", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

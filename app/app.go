// @title GeoAI App API
// @version 1.0
// @description This is the API documentation for GeoAI App.
// @termsOfService http://swagger.io/terms/

// @contact.name GeoAssistant Team
// @contact.email teysuizer1998@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

package app

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// For swagger use
	"geoai-app/controller"
	"geoai-app/docs"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

type App struct {
	DB     *sql.DB
	Routes *gin.Engine
}

func (a *App) CreateConnection() {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", UNAMEDB, PASSDB, HOSTDB, DBNAME)
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ { // Retry up to 5 times
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			if err = db.Ping(); err == nil {
				a.DB = db
				log.Println("Database connection established")
				return
			}
		}
		log.Printf("Retrying database connection... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to database: %v", err)
}

func (a *App) CreateRoutes() {
	routes := gin.Default()

	// Swagger documentation route
	docs.SwaggerInfo.BasePath = "/"
	// Register Swagger handler for UI and doc.json
	routes.GET("/swagger/*any", func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger/" {
			c.Redirect(302, "/swagger/index.html")
		} else {
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		}
	})

	// Initialize controllers
	userController := controller.NewUserController(a.DB)
	conversationController := controller.NewConversationController(a.DB)
	chatController := controller.NewChatController()

	// User routes
	routes.GET("/users", userController.GetUsers)
	routes.POST("/users", userController.CreateUser)

	// Conversation routes
	routes.GET("/conversations", conversationController.GetConversations)
	routes.POST("/conversations", conversationController.CreateConversation)

	// Chat route
	routes.POST("/chat", chatController.HandleChatRequest)

	a.Routes = routes
}

func (a *App) Run() {
	if err := a.Routes.Run(":8080"); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}

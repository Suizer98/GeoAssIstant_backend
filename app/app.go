package app

import (
 "database/sql"
 "fmt"
 "log"
 "time"

 "github.com/gin-gonic/gin"
 "geoai-app/controller"
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

	userController := controller.NewUserController(a.DB)
	conversationController := controller.NewConversationController(a.DB)

	// User routes
	routes.GET("/users", userController.GetUsers)
	routes.POST("/users", userController.CreateUser)

	// Conversation routes
	routes.GET("/conversations", conversationController.GetConversations)
	routes.POST("/conversations", conversationController.CreateConversation)

	a.Routes = routes
}

func (a *App) Run(){
 a.Routes.Run(":8080")
}
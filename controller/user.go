package controller

import (
	"database/sql"
	"net/http"
	"time"

	"geoai-app/model"
	"geoai-app/repository"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	DB *sql.DB
}

func NewUserController(db *sql.DB) *UserController {
	return &UserController{DB: db}
}

// @Summary Get all users
// @Description Retrieve all users or a specific user by ID
// @Tags users
// @Produce json
// @Param id query string false "User ID to fetch a specific user"
// @Success 200 {object} map[string]interface{} "Success response with users data"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Failed to fetch users"
// @Router /users [get]
func (u *UserController) GetUsers(ctx *gin.Context) {
	userID := ctx.Query("id")
	repo := repository.NewUserRepository(u.DB)

	if userID != "" {
		// Fetch specific user by ID
		user, err := repo.GetUserByID(userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to fetch user"})
			return
		}
		if user == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "User not found"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": user})
		return
	}

	// Fetch all users
	users, err := repo.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to fetch users"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": users})
}

// @Summary Create a new user
// @Description Add a new user to the database
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "New user data"
// @Success 201 {object} model.User "User successfully created"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Failed to create user"
// @Router /users [post]
func (u *UserController) CreateUser(ctx *gin.Context) {
	var req model.CreateUserRequest

	// Parse incoming JSON data
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "failed",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Create a new User object and set timestamps
	newUser := model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to the database
	repo := repository.NewUserRepository(u.DB)
	if err := repo.CreateUser(&newUser); err != nil {
		// Check for duplicate email error
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			ctx.JSON(http.StatusConflict, gin.H{
				"status":  "failed",
				"message": "Email already exists",
				"error":   err.Error(),
			})
			return
		}

		// Handle other errors
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "failed",
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	// Return the newly created user
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newUser})
}

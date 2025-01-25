package controller

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"geoai-app/model"
	"geoai-app/repository"
)

type UserController struct {
	DB *sql.DB
}

func NewUserController(db *sql.DB) *UserController {
	return &UserController{DB: db}
}

func (u *UserController) GetUsers(ctx *gin.Context) {
	userID := ctx.Query("id")
	repo := repository.NewUserRepository(u.DB)

	if userID != "" {
		// Fetch specific user by ID
		user, err := repo.GetUserById(userID)
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

func (u *UserController) CreateUser(ctx *gin.Context) {
	var newUser model.User
	if err := ctx.ShouldBindJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request data"})
		return
	}

	repo := repository.NewUserRepository(u.DB)
	if err := repo.CreateUser(&newUser); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newUser})
}

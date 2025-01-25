package repository

import (
	"database/sql"
	"geoai-app/model"
)

type UserRepositoryInterface interface {
	GetAllUsers() ([]model.User, error)
	GetUserByID(ID string) (*model.User, error)
	CreateUser(user *model.User) error
}

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepositoryInterface {
	return &UserRepository{DB: db}
}

// GetAllUsers retrieves all users from the database
func (r *UserRepository) GetAllUsers() ([]model.User, error) {
	rows, err := r.DB.Query("SELECT id, username, email, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(ID string) (*model.User, error) {
	row := r.DB.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1", ID)
	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(user *model.User) error {
	err := r.DB.QueryRow(
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		user.Username, user.Password, user.Email,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

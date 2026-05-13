package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Создание нового пользователя
func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания пользователя")
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("Пользователь успешно создан")

	return nil
}

// Поиск пользователя по email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at 
			  FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logrus.WithError(err).Error("Ошибка поиска пользователя по email")
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return user, nil
}

// Поиск пользователя по ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at 
			  FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logrus.WithError(err).Error("Ошибка поиска пользователя по ID")
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	return user, nil
}

// Проверка уникальности username
func (r *UserRepository) IsUsernameExists(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	var count int
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Получение email пользователя
func (r *UserRepository) GetUserEmail(userID string) string {
	query := `SELECT email FROM users WHERE id = $1`
	var email string
	err := r.db.QueryRow(query, userID).Scan(&email)
	if err != nil {
		logrus.WithError(err).Warn("Не удалось получить email пользователя")
		return ""
	}
	return email
}

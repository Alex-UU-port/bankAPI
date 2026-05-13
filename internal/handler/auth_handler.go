package handler

import (
	"bankAPI/internal/models"
	"bankAPI/internal/service"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithError(err).Warn("Ошибка декодирования JSON")
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Username == "" {
		http.Error(w, "Имя пользователя обязательно", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email обязателен", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Пароль должен содержать минимум 6 символов", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Регистрация успешна",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Вход в систему
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	token, userID, username, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.LoginResponse{
		Token:    token,
		Username: username,
		UserID:   userID,
	})
}

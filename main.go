package main

import (
	"bankAPI/internal/middleware"
	"bankAPI/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

var authService *service.AuthService

func main() {
	authService = service.NewAuthService()

	// Публичные маршруты
	http.HandleFunc("/", handleStart)
	http.HandleFunc("/any/", handleAnything)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)

	// Защищенные маршруты (с middleware)
	http.HandleFunc("/profile", middleware.AuthMiddleware(handleProfile))

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<h3>Стартовая страница<h3>"))
}

func handleAnything(w http.ResponseWriter, r *http.Request) {
	// Достаем метод (GET, POST и т.д.)
	method := r.Method

	// Достаем путь (/register, /login, /profile)
	path := r.URL.Path

	w.Write([]byte("Метод: " + method + ", Путь: " + path))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if err := authService.Register(req.Email, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "registered"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	token, err := authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	// Получаем user_id из контекста (установлен middleware)
	userID := r.Context().Value("user_id").(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"message": "This is protected endpoint",
	})
}

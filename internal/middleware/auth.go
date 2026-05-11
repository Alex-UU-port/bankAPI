package middleware

import (
	"context"
	"net/http"
)

// AuthMiddleware проверяет наличие токена в заголовке
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Берем токен из заголовка Authorization
		token := r.Header.Get("Authorization")

		// Если токена нет - ошибка
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// TODO: здесь будет проверка JWT
		// Пока просто пропускаем с фейковым user_id
		userID := "fake-user-id"

		// Кладем user_id в контекст
		ctx := context.WithValue(r.Context(), "user_id", userID)

		// Передаем управление следующему обработчику
		next(w, r.WithContext(ctx))
	}
}

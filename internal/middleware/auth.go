package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// Проверка JWT токена
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получение заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logrus.Warn("Отсутствует заголовок Authorization")
			http.Error(w, `{"error": "Требуется авторизация"}`, http.StatusUnauthorized)
			return
		}

		// Извлечение токена из Bearer схемы
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logrus.Warn("Неверный формат токена")
			http.Error(w, `{"error": "Неверный формат токена"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Парсинг и валидация токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			logrus.WithError(err).Warn("Ошибка валидации токена")
			http.Error(w, `{"error": "Недействительный токен"}`, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			logrus.Warn("Недействительный токен")
			http.Error(w, `{"error": "Недействительный токен"}`, http.StatusUnauthorized)
			return
		}

		// Извлечение user_id из claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Недействительный токен"}`, http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, `{"error": "Недействительный токен"}`, http.StatusUnauthorized)
			return
		}

		// Добавление user_id в контекст
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next(w, r.WithContext(ctx))
	}
}

// Логирование запросов
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("Входящий запрос")

		next(w, r)
	}
}

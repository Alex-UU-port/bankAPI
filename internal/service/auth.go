package service

import (
	"errors"
	"fmt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(email, password string) error {
	// Пока просто проверяем, что email не пустой
	if email == "" || password == "" {
		return errors.New("email and password required")
	}

	fmt.Printf("User registered: %s\n", email)
	return nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password required")
	}

	// Простой токен для теста
	token := fmt.Sprintf("fake-jwt-token-for-%s", email)
	return token, nil
}

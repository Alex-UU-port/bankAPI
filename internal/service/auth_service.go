package service

import (
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
	emailSvc  *EmailService
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, emailSvc *EmailService) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		emailSvc:  emailSvc,
	}
}

// Регистрация нового пользователя
func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	logrus.WithFields(logrus.Fields{
		"email":    req.Email,
		"username": req.Username,
	}).Info("Попытка регистрации нового пользователя")

	// Проверка существующего email
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Проверка существующего username
	exists, err := s.userRepo.IsUsernameExists(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("пользователь с таким именем уже существует")
	}

	// Хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Ошибка хеширования пароля")
		return nil, errors.New("ошибка при создании пользователя")
	}

	// Создание пользователя
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Отправка приветственного email
	s.emailSvc.SendWelcomeEmail(req.Email, req.Username)

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"email":    user.Email,
		"username": user.Username,
	}).Info("Пользователь успешно зарегистрирован")

	return user, nil
}

// Генерация JWT токена
func (s *AuthService) GenerateJWTToken(userID, email, username string) (string, error) {
	// Создаем MapClaims с необходимыми полями
	claims := jwt.MapClaims{
		"sub":      userID,
		"email":    email,
		"username": username,
		"iss":      "bank-api",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// Аутентификация пользователя
func (s *AuthService) Login(email, password string) (string, string, string, error) {
	logrus.WithField("email", email).Info("Попытка входа в систему")

	// Поиск пользователя
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", "", "", err
	}
	if user == nil {
		return "", "", "", errors.New("неверный email или пароль")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logrus.Warn("Неудачная попытка входа: неверный пароль")
		return "", "", "", errors.New("неверный email или пароль")
	}

	// Генерация JWT токена
	token, err := s.GenerateJWTToken(user.ID, user.Email, user.Username)
	if err != nil {
		logrus.WithError(err).Error("Ошибка генерации JWT токена")
		return "", "", "", errors.New("ошибка при входе в систему")
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"email":    user.Email,
		"username": user.Username,
	}).Info("Пользователь успешно вошел в систему")

	return token, user.ID, user.Username, nil
}

// Валидация JWT токена
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		logrus.WithError(err).Warn("Ошибка валидации токена")
		return "", errors.New("недействительный токен")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("недействительный токен: отсутствует user_id")
		}
		return userID, nil
	}

	return "", errors.New("недействительный токен")
}

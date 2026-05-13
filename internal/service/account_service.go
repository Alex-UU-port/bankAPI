package service

import (
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
	userRepo    *repository.UserRepository
}

func NewAccountService(accountRepo *repository.AccountRepository, userRepo *repository.UserRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		userRepo:    userRepo,
	}
}

// Создание нового счета
func (s *AccountService) CreateAccount(userID string, currency string) (*models.Account, error) {
	logrus.WithFields(logrus.Fields{
		"user_id":  userID,
		"currency": currency,
	}).Info("Создание нового банковского счета")

	// Генерация уникального номера счета
	accountNumber := generateAccountNumber()

	account := &models.Account{
		ID:            uuid.New().String(),
		UserID:        userID,
		AccountNumber: accountNumber,
		Balance:       0,
		Currency:      "RUB",
		CreatedAt:     time.Now(),
	}

	if err := s.accountRepo.Create(account); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"account_id":     account.ID,
		"account_number": account.AccountNumber,
		"user_id":        userID,
	}).Info("Банковский счет успешно создан")

	return account, nil
}

// Получение всех счетов пользователя
func (s *AccountService) GetUserAccounts(userID string) ([]*models.Account, error) {
	accounts, err := s.accountRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"accounts_count": len(accounts),
	}).Debug("Получены счета пользователя")

	return accounts, nil
}

// Получение счета по ID с проверкой прав
func (s *AccountService) GetAccountByID(accountID, userID string) (*models.Account, error) {
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("счет не найден")
	}

	// Проверка прав доступа
	if account.UserID != userID {
		logrus.Warn("Попытка доступа к чужому счету")
		return nil, fmt.Errorf("нет доступа к этому счету")
	}

	return account, nil
}

// Пополнение счета
func (s *AccountService) Deposit(accountID, userID string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть больше нуля")
	}

	// Проверка прав
	account, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return err
	}

	// Обновление баланса
	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"account_id":  accountID,
		"amount":      amount,
		"new_balance": newBalance,
		"user_id":     userID,
	}).Info("Счет успешно пополнен")

	return nil
}

// Списание со счета
func (s *AccountService) Withdraw(accountID, userID string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть больше нуля")
	}

	account, err := s.GetAccountByID(accountID, userID)
	if err != nil {
		return err
	}

	if account.Balance < amount {
		return fmt.Errorf("недостаточно средств на счете")
	}

	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"account_id":  accountID,
		"amount":      amount,
		"new_balance": newBalance,
	}).Info("Средства списаны со счета")

	return nil
}

// Генерация номера счета
func generateAccountNumber() string {
	prefix := "40817" // Префикс для счетов физлиц в рублях
	randomPart := generateRandomNumbers(15)
	return prefix + randomPart
}

// Генерация случайных чисел
func generateRandomNumbers(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%10]
	}
	return string(b)
}

package service

import (
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TransferService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	emailService    *EmailService
}

func NewTransferService(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	emailService *EmailService,
) *TransferService {
	return &TransferService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		emailService:    emailService,
	}
}

// Перевод между счетами
func (s *TransferService) Transfer(fromUserID, fromAccountID, toAccountNumber string, amount float64, description string) error {
	logrus.WithFields(logrus.Fields{
		"from_user_id":    fromUserID,
		"from_account_id": fromAccountID,
		"to_account":      toAccountNumber,
		"amount":          amount,
	}).Info("Инициирован перевод средств")

	if amount <= 0 {
		return fmt.Errorf("сумма перевода должна быть больше нуля")
	}

	// Получение счета отправителя
	fromAccount, err := s.accountRepo.FindByID(fromAccountID)
	if err != nil {
		return err
	}
	if fromAccount == nil {
		return fmt.Errorf("счет отправителя не найден")
	}
	if fromAccount.UserID != fromUserID {
		return fmt.Errorf("нет доступа к этому счету")
	}

	// Проверка достаточности средств
	if fromAccount.Balance < amount {
		return fmt.Errorf("недостаточно средств на счете")
	}

	// Получение счета получателя
	toAccount, err := s.accountRepo.FindByNumber(toAccountNumber)
	if err != nil {
		return err
	}
	if toAccount == nil {
		return fmt.Errorf("счет получателя не найден")
	}

	// Выполнение перевода
	if err := s.executeTransfer(fromAccount, toAccount, amount, description); err != nil {
		return err
	}

	return nil
}

// Выполнение перевода (внутренний метод)
func (s *TransferService) executeTransfer(fromAccount, toAccount *models.Account, amount float64, description string) error {
	// Начинаем транзакцию
	tx, err := s.accountRepo.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновление балансов в рамках транзакции
	newFromBalance := fromAccount.Balance - amount
	if err := s.accountRepo.UpdateBalanceWithTx(tx, fromAccount.ID, newFromBalance); err != nil {
		return err
	}

	newToBalance := toAccount.Balance + amount
	if err := s.accountRepo.UpdateBalanceWithTx(tx, toAccount.ID, newToBalance); err != nil {
		return err
	}

	// Создание записи транзакции
	transaction := &models.Transaction{
		ID:            uuid.New().String(),
		FromAccountID: &fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        amount,
		Type:          "transfer",
		Status:        "completed",
		Description:   description,
		CreatedAt:     time.Now(),
	}

	if err := s.transactionRepo.CreateWithTx(tx, transaction); err != nil {
		return err
	}

	// Фиксация транзакции
	if err := tx.Commit(); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"from_account": fromAccount.AccountNumber,
		"to_account":   toAccount.AccountNumber,
		"amount":       amount,
	}).Info("Перевод успешно выполнен")

	// Отправка уведомлений (асинхронно)
	go s.sendNotifications(fromAccount.UserID, toAccount.UserID, amount)

	return nil
}

// Отправка уведомлений о переводе
func (s *TransferService) sendNotifications(fromUserID, toUserID string, amount float64) {
	// Здесь можно отправить email уведомления
	logrus.WithFields(logrus.Fields{
		"from_user": fromUserID,
		"to_user":   toUserID,
		"amount":    amount,
	}).Debug("Отправка уведомлений о переводе")
}

// getUserEmail получает email пользователя
func (s *TransferService) getUserEmail(userID string) string {
	// Этот метод должен быть реализован через UserRepository
	// Для краткости возвращаем заглушку
	return ""
}

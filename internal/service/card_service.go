package service

import (
	"bankAPI/internal/crypto"
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type CardService struct {
	cardRepo     *repository.CardRepository
	accountRepo  *repository.AccountRepository
	emailService *EmailService
}

func NewCardService(
	cardRepo *repository.CardRepository,
	accountRepo *repository.AccountRepository,
	emailService *EmailService,
) *CardService {
	return &CardService{
		cardRepo:     cardRepo,
		accountRepo:  accountRepo,
		emailService: emailService,
	}
}

// Создание новой карты
func (s *CardService) CreateCard(accountID, userID string) (*models.Card, string, error) {
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Запрос на выпуск новой карты")

	// Проверка прав доступа к счету
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, "", err
	}
	if account == nil {
		return nil, "", fmt.Errorf("счет не найден")
	}
	if account.UserID != userID {
		logrus.Warn("Попытка выпустить карту на чужой счет")
		return nil, "", fmt.Errorf("нет доступа к этому счету")
	}

	// Генерация данных карты
	cardNumber := crypto.GenerateCardNumber()
	cvv := crypto.GenerateCVV()
	expiry := crypto.GenerateExpiry()

	// Шифрование номера карты и срока действия
	encryptedNumber, err := crypto.EncryptData(cardNumber)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка шифрования номера карты: %w", err)
	}

	encryptedExpiry, err := crypto.EncryptData(expiry)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка шифрования срока действия: %w", err)
	}

	// Хеширование CVV
	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка хеширования CVV: %w", err)
	}

	// Маскированный номер для отображения
	maskedNumber := cardNumber[:4] + " **** **** " + cardNumber[len(cardNumber)-4:]

	// HMAC для проверки целостности
	dataForHMAC := cardNumber + expiry
	hmacSignature := crypto.GenerateHMAC(dataForHMAC)

	card := &models.Card{
		ID:                  uuid.New().String(),
		AccountID:           accountID,
		CardNumberEncrypted: encryptedNumber,
		CardNumberMasked:    maskedNumber,
		ExpiryEncrypted:     encryptedExpiry,
		ExpiryMasked:        expiry,
		CVVHash:             string(cvvHash),
		HMACSignature:       hmacSignature,
		IsActive:            true,
		CreatedAt:           time.Now(),
	}

	if err := s.cardRepo.Create(card); err != nil {
		return nil, "", err
	}

	logrus.WithFields(logrus.Fields{
		"card_id":       card.ID,
		"account_id":    accountID,
		"masked_number": maskedNumber,
	}).Info("Карта успешно выпущена")

	// Отправка email уведомления
	go s.sendCardNotification(account.UserID, maskedNumber)

	return card, cvv, nil
}

// Получение карт пользователя
func (s *CardService) GetUserCards(accountID, userID string) ([]*models.Card, error) {
	// Проверка прав доступа
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("счет не найден")
	}
	if account.UserID != userID {
		return nil, fmt.Errorf("нет доступа к этому счету")
	}

	cards, err := s.cardRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":     userID,
		"account_id":  accountID,
		"cards_count": len(cards),
	}).Debug("Получены карты пользователя")

	return cards, nil
}

// Оплата картой
func (s *CardService) ProcessPayment(cardID string, cvv string, amount float64) error {
	logrus.WithFields(logrus.Fields{
		"card_id": cardID,
		"amount":  amount,
	}).Info("Обработка платежа по карте")

	if amount <= 0 {
		return fmt.Errorf("сумма платежа должна быть больше нуля")
	}

	// Получение карты
	card, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		return err
	}
	if card == nil {
		return fmt.Errorf("карта не найдена")
	}

	if !card.IsActive {
		return fmt.Errorf("карта заблокирована")
	}

	// Проверка CVV
	if err := bcrypt.CompareHashAndPassword([]byte(card.CVVHash), []byte(cvv)); err != nil {
		logrus.Warn("Неверный CVV код")
		return fmt.Errorf("неверный CVV код")
	}

	// Проверка целостности данных (HMAC)
	// Для проверки нужно расшифровать номер и срок
	decryptedNumber, err := crypto.DecryptData(card.CardNumberEncrypted)
	if err != nil {
		return fmt.Errorf("ошибка расшифровки данных карты")
	}

	decryptedExpiry, err := crypto.DecryptData(card.ExpiryEncrypted)
	if err != nil {
		return fmt.Errorf("ошибка расшифровки данных карты")
	}

	dataForHMAC := decryptedNumber + decryptedExpiry
	if !crypto.VerifyHMAC(dataForHMAC, card.HMACSignature) {
		logrus.Error("Нарушена целостность данных карты")
		return fmt.Errorf("ошибка проверки целостности данных карты")
	}

	// Получение счета
	account, err := s.accountRepo.FindByID(card.AccountID)
	if err != nil {
		return err
	}

	// Проверка достаточности средств
	if account.Balance < amount {
		return fmt.Errorf("недостаточно средств на счете")
	}

	// Списание средств
	newBalance := account.Balance - amount
	if err := s.accountRepo.UpdateBalance(account.ID, newBalance); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"card_id":     cardID,
		"account_id":  account.ID,
		"amount":      amount,
		"new_balance": newBalance,
	}).Info("Платеж по карте успешно выполнен")

	// Отправка уведомления
	go s.sendPaymentNotification(account.UserID, amount, card.CardNumberMasked)

	return nil
}

// Отправка уведомления о выпуске карты
func (s *CardService) sendCardNotification(userID, maskedNumber string) {
	email := s.getUserEmail(userID)
	if email != "" {
		subject := "Выпущена новая банковская карта"
		body := fmt.Sprintf("На ваше имя выпущена карта %s. Храните данные карты в безопасности!", maskedNumber)
		s.emailService.SendEmail(email, subject, body)
	}
}

// Отправка уведомления о платеже
func (s *CardService) sendPaymentNotification(userID string, amount float64, maskedNumber string) {
	email := s.getUserEmail(userID)
	if email != "" {
		subject := "Совершен платеж по карте"
		body := fmt.Sprintf("С карты %s списано %.2f RUB", maskedNumber, amount)
		s.emailService.SendEmail(email, subject, body)
	}
}

func (s *CardService) getUserEmail(userID string) string {
	// Временная заглушка
	return ""
}

package service

import (
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type CreditService struct {
	creditRepo   *repository.CreditRepository
	accountRepo  *repository.AccountRepository
	scheduleRepo *repository.ScheduleRepository
	emailService *EmailService
	cbrClient    *CBRClient
}

func NewCreditService(
	creditRepo *repository.CreditRepository,
	accountRepo *repository.AccountRepository,
	scheduleRepo *repository.ScheduleRepository,
	emailService *EmailService,
	cbrClient *CBRClient,
) *CreditService {
	return &CreditService{
		creditRepo:   creditRepo,
		accountRepo:  accountRepo,
		scheduleRepo: scheduleRepo,
		emailService: emailService,
		cbrClient:    cbrClient,
	}
}

// Оформление кредита
func (s *CreditService) CreateCredit(userID string, req models.CreateCreditRequest) (*models.Credit, []*models.PaymentSchedule, error) {
	logrus.WithFields(logrus.Fields{
		"user_id":     userID,
		"account_id":  req.AccountID,
		"amount":      req.Amount,
		"term_months": req.TermMonths,
	}).Info("Оформление кредита")

	// Проверка прав доступа к счету
	account, err := s.accountRepo.FindByID(req.AccountID)
	if err != nil {
		return nil, nil, err
	}
	if account == nil {
		return nil, nil, fmt.Errorf("счет не найден")
	}
	if account.UserID != userID {
		return nil, nil, fmt.Errorf("нет доступа к этому счету")
	}

	// Получение ключевой ставки ЦБ РФ
	keyRate, err := s.cbrClient.GetKeyRate()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось получить ключевую ставку, используется значение по умолчанию")
		keyRate = 16.0 // Значение по умолчанию
	}

	// Базовая ставка = ключевая ставка + 2%
	interestRate := keyRate + 2.0

	// Расчет аннуитетного платежа
	monthlyRate := interestRate / 100 / 12
	monthlyPayment := req.Amount * monthlyRate * math.Pow(1+monthlyRate, float64(req.TermMonths)) /
		(math.Pow(1+monthlyRate, float64(req.TermMonths)) - 1)

	credit := &models.Credit{
		ID:              uuid.New().String(),
		UserID:          userID,
		AccountID:       req.AccountID,
		Amount:          req.Amount,
		InterestRate:    interestRate,
		TermMonths:      req.TermMonths,
		MonthlyPayment:  monthlyPayment,
		RemainingAmount: req.Amount,
		Status:          "active",
		IssuedAt:        time.Now(),
	}

	// Создание графика платежей
	schedules := s.generatePaymentSchedule(credit, monthlyRate)

	// Сохранение кредита и графика в БД (транзакция)
	tx, err := s.creditRepo.BeginTx()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	if err := s.creditRepo.CreateWithTx(tx, credit); err != nil {
		return nil, nil, err
	}

	if err := s.scheduleRepo.CreateBatchWithTx(tx, schedules); err != nil {
		return nil, nil, err
	}

	// Зачисление суммы кредита на счет
	newBalance := account.Balance + req.Amount
	if err := s.accountRepo.UpdateBalanceWithTx(tx, account.ID, newBalance); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	logrus.WithFields(logrus.Fields{
		"credit_id":       credit.ID,
		"amount":          credit.Amount,
		"interest_rate":   credit.InterestRate,
		"monthly_payment": credit.MonthlyPayment,
		"key_rate":        keyRate,
	}).Info("Кредит успешно оформлен")

	// Отправка email уведомления
	go s.sendCreditNotification(account.UserID, credit)

	return credit, schedules, nil
}

// Генерация графика аннуитетных платежей
func (s *CreditService) generatePaymentSchedule(credit *models.Credit, monthlyRate float64) []*models.PaymentSchedule {
	schedules := make([]*models.PaymentSchedule, credit.TermMonths)
	remaining := credit.Amount

	for i := 0; i < credit.TermMonths; i++ {
		interest := remaining * monthlyRate
		principal := credit.MonthlyPayment - interest
		if principal > remaining {
			principal = remaining
		}
		remaining -= principal

		schedules[i] = &models.PaymentSchedule{
			ID:              uuid.New().String(),
			CreditID:        credit.ID,
			PaymentNumber:   i + 1,
			DueDate:         credit.IssuedAt.AddDate(0, i+1, 0),
			Amount:          credit.MonthlyPayment,
			PrincipalAmount: principal,
			InterestAmount:  interest,
			PenaltyAmount:   0,
			Status:          "pending",
		}
	}

	return schedules
}

// Получение графика платежей по кредиту
func (s *CreditService) GetPaymentSchedule(creditID, userID string) ([]*models.PaymentSchedule, error) {
	// Проверка прав доступа
	credit, err := s.creditRepo.FindByID(creditID)
	if err != nil {
		return nil, err
	}
	if credit == nil {
		return nil, fmt.Errorf("кредит не найден")
	}
	if credit.UserID != userID {
		return nil, fmt.Errorf("нет доступа к этому кредиту")
	}

	schedules, err := s.scheduleRepo.FindByCreditID(creditID)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// Обработка просроченных платежей (вызывается шедулером)
func (s *CreditService) ProcessOverduePayments() {
	logrus.Info("Запуск обработки просроченных платежей")

	overdueSchedules, err := s.scheduleRepo.FindOverduePending()
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения просроченных платежей")
		return
	}

	for _, schedule := range overdueSchedules {
		s.processOverdueSchedule(schedule)
	}

	logrus.Infof("Обработано просроченных платежей: %d", len(overdueSchedules))
}

// Обработка одного просроченного платежа
func (s *CreditService) processOverdueSchedule(schedule *models.PaymentSchedule) {
	logrus.WithFields(logrus.Fields{
		"schedule_id":    schedule.ID,
		"credit_id":      schedule.CreditID,
		"payment_number": schedule.PaymentNumber,
	}).Info("Обработка просроченного платежа")

	// Получение кредита
	credit, err := s.creditRepo.FindByID(schedule.CreditID)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения кредита")
		return
	}

	// Получение счета
	account, err := s.accountRepo.FindByID(credit.AccountID)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения счета")
		return
	}

	totalAmount := schedule.Amount + schedule.PenaltyAmount

	if account.Balance >= totalAmount {
		// Списание средств
		newBalance := account.Balance - totalAmount
		if err := s.accountRepo.UpdateBalance(account.ID, newBalance); err != nil {
			logrus.WithError(err).Error("Ошибка списания средств")
			return
		}

		// Обновление статуса платежа
		now := time.Now()
		schedule.Status = "paid"
		schedule.PaidAt = &now
		if err := s.scheduleRepo.Update(schedule); err != nil {
			logrus.WithError(err).Error("Ошибка обновления статуса платежа")
		}

		// Обновление остатка по кредиту
		credit.RemainingAmount -= schedule.PrincipalAmount
		if credit.RemainingAmount <= 0 {
			credit.Status = "closed"
		}
		if err := s.creditRepo.Update(credit); err != nil {
			logrus.WithError(err).Error("Ошибка обновления кредита")
		}

		// Отправка уведомления
		go s.sendPaymentNotification(credit.UserID, schedule)

		logrus.WithFields(logrus.Fields{
			"schedule_id": schedule.ID,
			"amount":      totalAmount,
			"penalty":     schedule.PenaltyAmount,
		}).Info("Просроченный платеж списан")

	} else {
		// Начисление штрафа (+10% к сумме платежа)
		newPenalty := schedule.Amount * 0.1
		schedule.PenaltyAmount += newPenalty
		schedule.Status = "overdue"

		if err := s.scheduleRepo.Update(schedule); err != nil {
			logrus.WithError(err).Error("Ошибка начисления штрафа")
		}

		logrus.WithFields(logrus.Fields{
			"schedule_id":   schedule.ID,
			"new_penalty":   newPenalty,
			"total_penalty": schedule.PenaltyAmount,
		}).Warn("Начислен штраф за просрочку")
	}
}

// Отправка уведомления о кредите
func (s *CreditService) sendCreditNotification(userID string, credit *models.Credit) {
	email := s.getUserEmail(userID)
	if email != "" {
		subject := "Кредит оформлен"
		body := fmt.Sprintf(
			"Вам одобрен кредит на сумму %.2f RUB.\n"+
				"Процентная ставка: %.2f%%\n"+
				"Ежемесячный платеж: %.2f RUB\n"+
				"Срок: %d месяцев",
			credit.Amount, credit.InterestRate, credit.MonthlyPayment, credit.TermMonths,
		)
		s.emailService.SendEmail(email, subject, body)
	}
}

// Отправка уведомления о платеже по кредиту
func (s *CreditService) sendPaymentNotification(userID string, schedule *models.PaymentSchedule) {
	email := s.getUserEmail(userID)
	if email != "" {
		subject := "Платеж по кредиту"
		body := fmt.Sprintf(
			"Совершен платеж по кредиту №%d.\n"+
				"Сумма: %.2f RUB\n"+
				"Из них: основной долг %.2f RUB, проценты %.2f RUB",
			schedule.PaymentNumber, schedule.Amount, schedule.PrincipalAmount, schedule.InterestAmount,
		)
		if schedule.PenaltyAmount > 0 {
			body += fmt.Sprintf("\nШтраф: %.2f RUB", schedule.PenaltyAmount)
		}
		s.emailService.SendEmail(email, subject, body)
	}
}

func (s *CreditService) getUserEmail(userID string) string {
	// Временная заглушка
	return ""
}

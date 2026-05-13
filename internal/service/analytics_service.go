package service

import (
	"bankAPI/internal/models"
	"bankAPI/internal/repository"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type AnalyticsService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	creditRepo      *repository.CreditRepository
	scheduleRepo    *repository.ScheduleRepository
	cbrClient       *CBRClient
}

func NewAnalyticsService(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	creditRepo *repository.CreditRepository,
	scheduleRepo *repository.ScheduleRepository,
	cbrClient *CBRClient,
) *AnalyticsService {
	return &AnalyticsService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		creditRepo:      creditRepo,
		scheduleRepo:    scheduleRepo,
		cbrClient:       cbrClient,
	}
}

// Получение полной аналитики для пользователя
func (s *AnalyticsService) GetAnalytics(userID string) (*models.AnalyticsResponse, error) {
	logrus.WithField("user_id", userID).Info("Запрос аналитики")

	// Получение всех счетов пользователя
	accounts, err := s.accountRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Получение ключевой ставки ЦБ
	keyRate, err := s.cbrClient.GetKeyRate()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось получить ключевую ставку")
		keyRate = 16.0
	}

	// Статистика за последний месяц
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	var totalIncome, totalExpense float64
	monthlyStats := make(map[string]float64)

	for _, account := range accounts {
		transactions, err := s.transactionRepo.FindByAccountIDAndPeriod(
			account.ID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
		)
		if err != nil {
			logrus.WithError(err).Warn("Ошибка получения транзакций для счета")
			continue
		}

		for _, tx := range transactions {
			// Доходы (пополнения и зачисления)
			if tx.ToAccountID == account.ID && tx.Type != "transfer" {
				totalIncome += tx.Amount
			}
			// Расходы (списания и переводы)
			if tx.FromAccountID != nil && *tx.FromAccountID == account.ID {
				totalExpense += tx.Amount
			}

			// Статистика по дням
			dayKey := tx.CreatedAt.Format("2006-01-02")
			monthlyStats[dayKey] += tx.Amount
		}
	}

	// Расчет кредитной нагрузки
	credits, err := s.creditRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	var totalMonthlyPayment float64
	for _, credit := range credits {
		totalMonthlyPayment += credit.MonthlyPayment
	}

	// Общий баланс всех счетов
	var totalBalance float64
	for _, account := range accounts {
		totalBalance += account.Balance
	}

	// Кредитная нагрузка (отношение платежей к доходу)
	var creditLoad float64
	if totalIncome > 0 {
		creditLoad = (totalMonthlyPayment / totalIncome) * 100
	}

	analytics := &models.AnalyticsResponse{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Balance:      totalBalance,
		CreditLoad:   creditLoad,
		MonthlyStats: monthlyStats,
		KeyRate:      keyRate,
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"balance":       totalBalance,
		"credit_load":   creditLoad,
	}).Info("Аналитика успешно сформирована")

	return analytics, nil
}

// Прогноз баланса на N дней
func (s *AnalyticsService) PredictBalance(accountID, userID string, days int) ([]map[string]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": accountID,
		"days":       days,
	}).Info("Запрос прогноза баланса")

	if days > 365 {
		return nil, fmt.Errorf("максимальный период прогноза - 365 дней")
	}

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

	// Получение истории транзакций за последние 30 дней
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	transactions, err := s.transactionRepo.FindByAccountIDAndPeriod(
		accountID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}

	// Расчет среднего дневного расхода и дохода
	var avgDailyExpense, avgDailyIncome float64
	var expenseDays, incomeDays int

	for _, tx := range transactions {
		if tx.FromAccountID != nil && *tx.FromAccountID == accountID {
			avgDailyExpense += tx.Amount
			expenseDays++
		}
		if tx.ToAccountID == accountID {
			avgDailyIncome += tx.Amount
			incomeDays++
		}
	}

	if expenseDays > 0 {
		avgDailyExpense = avgDailyExpense / float64(expenseDays)
	}
	if incomeDays > 0 {
		avgDailyIncome = avgDailyIncome / float64(incomeDays)
	}

	// Получение предстоящих платежей по кредитам
	upcomingPayments, err := s.getUpcomingPayments(userID, days)
	if err != nil {
		logrus.WithError(err).Warn("Ошибка получения предстоящих платежей")
	}

	// Формирование прогноза
	predictions := make([]map[string]interface{}, days)
	currentBalance := account.Balance

	for i := 0; i < days; i++ {
		date := endDate.AddDate(0, 0, i+1)

		// Начинаем с текущего баланса
		dailyChange := avgDailyIncome - avgDailyExpense

		// Проверяем, есть ли платеж по кредиту в этот день
		paymentAmount := s.checkScheduledPayment(upcomingPayments, date)
		if paymentAmount > 0 {
			dailyChange -= paymentAmount
		}

		currentBalance += dailyChange
		if currentBalance < 0 {
			currentBalance = 0
		}

		predictions[i] = map[string]interface{}{
			"day":              i + 1,
			"date":             date.Format("2006-01-02"),
			"balance":          roundToTwoDecimals(currentBalance),
			"predicted_change": roundToTwoDecimals(dailyChange),
		}
	}

	logrus.WithFields(logrus.Fields{
		"account_id":    accountID,
		"days":          days,
		"final_balance": predictions[days-1]["balance"],
	}).Info("Прогноз баланса успешно сформирован")

	return predictions, nil
}

// Получение предстоящих платежей по кредитам
func (s *AnalyticsService) getUpcomingPayments(userID string, days int) ([]*models.PaymentSchedule, error) {
	credits, err := s.creditRepo.FindActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	var allPayments []*models.PaymentSchedule
	endDate := time.Now().AddDate(0, 0, days)

	for _, credit := range credits {
		payments, err := s.scheduleRepo.FindByCreditIDAndDateRange(
			credit.ID,
			time.Now(),
			endDate,
		)
		if err != nil {
			continue
		}
		allPayments = append(allPayments, payments...)
	}

	return allPayments, nil
}

// Проверка наличия платежа в указанную дату
func (s *AnalyticsService) checkScheduledPayment(payments []*models.PaymentSchedule, date time.Time) float64 {
	for _, payment := range payments {
		if payment.DueDate.Year() == date.Year() &&
			payment.DueDate.Month() == date.Month() &&
			payment.DueDate.Day() == date.Day() {
			return payment.Amount + payment.PenaltyAmount
		}
	}
	return 0
}

// Округление до двух знаков
func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100)) / 100
}

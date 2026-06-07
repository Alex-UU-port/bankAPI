package scheduler

import (
	"bankAPI/internal/service"
	"time"

	"github.com/sirupsen/logrus"
)

// PaymentScheduler - шедулер для обработки платежей по кредитам
type PaymentScheduler struct {
	creditService *service.CreditService
	ticker        *time.Ticker
	stopChan      chan bool
}

// NewPaymentScheduler создает новый шедулер
func NewPaymentScheduler(creditService *service.CreditService) *PaymentScheduler {
	return &PaymentScheduler{
		creditService: creditService,
		stopChan:      make(chan bool),
	}
}

// Start запускает шедулер (каждые N часов)
func (s *PaymentScheduler) Start(intervalHours int) {
	interval := time.Duration(intervalHours) * time.Hour
	s.ticker = time.NewTicker(interval)

	logrus.WithField("interval_hours", intervalHours).Info("Шедулер платежей запущен")

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.processOverduePayments()
			case <-s.stopChan:
				s.ticker.Stop()
				logrus.Info("Шедулер платежей остановлен")
				return
			}
		}
	}()
}

// Stop останавливает шедулер
func (s *PaymentScheduler) Stop() {
	s.stopChan <- true
}

// processOverduePayments обрабатывает просроченные платежи
func (s *PaymentScheduler) processOverduePayments() {
	logrus.Info("Запуск обработки просроченных платежей")

	startTime := time.Now()

	// Вызываем метод сервиса кредитов
	s.creditService.ProcessOverduePayments()

	duration := time.Since(startTime)
	logrus.WithField("duration_ms", duration.Milliseconds()).Info("Обработка просроченных платежей завершена")
}

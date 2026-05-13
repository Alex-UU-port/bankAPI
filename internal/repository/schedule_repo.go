package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type ScheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Создание графика платежей
func (r *ScheduleRepository) Create(schedule *models.PaymentSchedule) error {
	query := `INSERT INTO payment_schedules (id, credit_id, payment_number, due_date, amount, 
              principal_amount, interest_amount, penalty_amount, status) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(query, schedule.ID, schedule.CreditID, schedule.PaymentNumber,
		schedule.DueDate, schedule.Amount, schedule.PrincipalAmount,
		schedule.InterestAmount, schedule.PenaltyAmount, schedule.Status)

	if err != nil {
		logrus.WithError(err).Error("Ошибка создания графика платежей")
		return fmt.Errorf("ошибка создания графика: %w", err)
	}

	return nil
}

// Массовое создание графика платежей с транзакцией
func (r *ScheduleRepository) CreateBatchWithTx(tx *sql.Tx, schedules []*models.PaymentSchedule) error {
	query := `INSERT INTO payment_schedules (id, credit_id, payment_number, due_date, amount, 
              principal_amount, interest_amount, penalty_amount, status) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, schedule := range schedules {
		_, err := tx.Exec(query, schedule.ID, schedule.CreditID, schedule.PaymentNumber,
			schedule.DueDate, schedule.Amount, schedule.PrincipalAmount,
			schedule.InterestAmount, schedule.PenaltyAmount, schedule.Status)
		if err != nil {
			return err
		}
	}

	return nil
}

// Поиск графика по ID кредита
func (r *ScheduleRepository) FindByCreditID(creditID string) ([]*models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, payment_number, due_date, amount, principal_amount, 
              interest_amount, penalty_amount, status, paid_at 
              FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_number`

	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения графика: %w", err)
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		schedule := &models.PaymentSchedule{}
		err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentNumber,
			&schedule.DueDate, &schedule.Amount, &schedule.PrincipalAmount,
			&schedule.InterestAmount, &schedule.PenaltyAmount,
			&schedule.Status, &schedule.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Поиск просроченных платежей
func (r *ScheduleRepository) FindOverduePending() ([]*models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, payment_number, due_date, amount, principal_amount, 
              interest_amount, penalty_amount, status, paid_at 
              FROM payment_schedules 
              WHERE status = 'pending' AND due_date < CURRENT_DATE`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения просроченных платежей: %w", err)
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		schedule := &models.PaymentSchedule{}
		err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentNumber,
			&schedule.DueDate, &schedule.Amount, &schedule.PrincipalAmount,
			&schedule.InterestAmount, &schedule.PenaltyAmount,
			&schedule.Status, &schedule.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Поиск платежей по кредиту в диапазоне дат
func (r *ScheduleRepository) FindByCreditIDAndDateRange(creditID string, startDate, endDate time.Time) ([]*models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, payment_number, due_date, amount, principal_amount, 
              interest_amount, penalty_amount, status, paid_at 
              FROM payment_schedules 
              WHERE credit_id = $1 AND due_date BETWEEN $2 AND $3
              ORDER BY due_date`

	rows, err := r.db.Query(query, creditID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения платежей: %w", err)
	}
	defer rows.Close()

	var schedules []*models.PaymentSchedule
	for rows.Next() {
		schedule := &models.PaymentSchedule{}
		err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentNumber,
			&schedule.DueDate, &schedule.Amount, &schedule.PrincipalAmount,
			&schedule.InterestAmount, &schedule.PenaltyAmount,
			&schedule.Status, &schedule.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Обновление платежа
func (r *ScheduleRepository) Update(schedule *models.PaymentSchedule) error {
	query := `UPDATE payment_schedules SET status = $1, paid_at = $2, penalty_amount = $3 
              WHERE id = $4`
	_, err := r.db.Exec(query, schedule.Status, schedule.PaidAt, schedule.PenaltyAmount, schedule.ID)
	return err
}

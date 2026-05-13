package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

// Создание кредита
func (r *CreditRepository) Create(credit *models.Credit) error {
	query := `INSERT INTO credits (id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(query, credit.ID, credit.UserID, credit.AccountID, credit.Amount,
		credit.InterestRate, credit.TermMonths, credit.MonthlyPayment,
		credit.RemainingAmount, credit.Status, credit.IssuedAt)

	if err != nil {
		logrus.WithError(err).Error("Ошибка создания кредита")
		return fmt.Errorf("ошибка создания кредита: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"credit_id": credit.ID,
		"user_id":   credit.UserID,
		"amount":    credit.Amount,
	}).Info("Кредит успешно создан")

	return nil
}

// Создание кредита с транзакцией
func (r *CreditRepository) CreateWithTx(tx *sql.Tx, credit *models.Credit) error {
	query := `INSERT INTO credits (id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := tx.Exec(query, credit.ID, credit.UserID, credit.AccountID, credit.Amount,
		credit.InterestRate, credit.TermMonths, credit.MonthlyPayment,
		credit.RemainingAmount, credit.Status, credit.IssuedAt)

	return err
}

// Поиск кредита по ID
func (r *CreditRepository) FindByID(id string) (*models.Credit, error) {
	query := `SELECT id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at 
              FROM credits WHERE id = $1`

	credit := &models.Credit{}
	err := r.db.QueryRow(query, id).Scan(
		&credit.ID, &credit.UserID, &credit.AccountID, &credit.Amount,
		&credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment,
		&credit.RemainingAmount, &credit.Status, &credit.IssuedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска кредита: %w", err)
	}

	return credit, nil
}

// Поиск активных кредитов пользователя
func (r *CreditRepository) FindActiveByUserID(userID string) ([]*models.Credit, error) {
	query := `SELECT id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at 
              FROM credits WHERE user_id = $1 AND status = 'active' ORDER BY issued_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения кредитов: %w", err)
	}
	defer rows.Close()

	var credits []*models.Credit
	for rows.Next() {
		credit := &models.Credit{}
		err := rows.Scan(&credit.ID, &credit.UserID, &credit.AccountID, &credit.Amount,
			&credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment,
			&credit.RemainingAmount, &credit.Status, &credit.IssuedAt)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}

	return credits, nil
}

// Обновление кредита
func (r *CreditRepository) Update(credit *models.Credit) error {
	query := `UPDATE credits SET remaining_amount = $1, status = $2 WHERE id = $3`
	_, err := r.db.Exec(query, credit.RemainingAmount, credit.Status, credit.ID)
	return err
}

// Начало транзакции
func (r *CreditRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// Обновление баланса счета в транзакции
func (r *CreditRepository) UpdateBalanceWithTx(tx *sql.Tx, accountID string, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := tx.Exec(query, newBalance, accountID)
	return err
}

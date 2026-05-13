package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Создание счета
func (r *AccountRepository) Create(account *models.Account) error {
	query := `INSERT INTO accounts (id, user_id, account_number, balance, currency, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, account.ID, account.UserID, account.AccountNumber,
		account.Balance, account.Currency, account.CreatedAt)

	if err != nil {
		logrus.WithError(err).Error("Ошибка создания счета")
		return fmt.Errorf("ошибка создания счета: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"account_id":     account.ID,
		"account_number": account.AccountNumber,
		"user_id":        account.UserID,
	}).Info("Счет успешно создан")

	return nil
}

// Получение счета по ID
func (r *AccountRepository) FindByID(id string) (*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
			  FROM accounts WHERE id = $1`

	account := &models.Account{}
	err := r.db.QueryRow(query, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber,
		&account.Balance, &account.Currency, &account.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения счета: %w", err)
	}

	return account, nil
}

// Получение всех счетов пользователя
func (r *AccountRepository) FindByUserID(userID string) ([]*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
			  FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения счетов: %w", err)
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		account := &models.Account{}
		err := rows.Scan(&account.ID, &account.UserID, &account.AccountNumber,
			&account.Balance, &account.Currency, &account.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Поиск счета по номеру
func (r *AccountRepository) FindByNumber(accountNumber string) (*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
			  FROM accounts WHERE account_number = $1`

	account := &models.Account{}
	err := r.db.QueryRow(query, accountNumber).Scan(
		&account.ID, &account.UserID, &account.AccountNumber,
		&account.Balance, &account.Currency, &account.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска счета: %w", err)
	}

	return account, nil
}

// Обновление баланса счета
func (r *AccountRepository) UpdateBalance(accountID string, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`

	result, err := r.db.Exec(query, newBalance, accountID)
	if err != nil {
		return fmt.Errorf("ошибка обновления баланса: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("счет не найден")
	}

	logrus.WithFields(logrus.Fields{
		"account_id":  accountID,
		"new_balance": newBalance,
	}).Debug("Баланс счета обновлен")

	return nil
}

// Начало транзакции
func (r *AccountRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// Обновление баланса с транзакцией
func (r *AccountRepository) UpdateBalanceWithTx(tx *sql.Tx, accountID string, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := tx.Exec(query, newBalance, accountID)
	return err
}

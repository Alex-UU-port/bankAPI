package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Создание транзакции
func (r *TransactionRepository) Create(tx *models.Transaction) error {
	query := `INSERT INTO transactions (id, from_account_id, to_account_id, amount, type, status, description, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query, tx.ID, tx.FromAccountID, tx.ToAccountID,
		tx.Amount, tx.Type, tx.Status, tx.Description, tx.CreatedAt)

	if err != nil {
		logrus.WithError(err).Error("Ошибка создания транзакции")
		return fmt.Errorf("ошибка создания транзакции: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"transaction_id": tx.ID,
		"amount":         tx.Amount,
		"type":           tx.Type,
	}).Info("Транзакция создана")

	return nil
}

// Получение всех транзакций по счету
func (r *TransactionRepository) FindByAccountID(accountID string) ([]*models.Transaction, error) {
	query := `SELECT id, from_account_id, to_account_id, amount, type, status, description, created_at 
			  FROM transactions 
			  WHERE from_account_id = $1 OR to_account_id = $1 
			  ORDER BY created_at DESC
			  LIMIT 100`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения транзакций: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount,
			&tx.Type, &tx.Status, &tx.Description, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// Получение транзакций за период
func (r *TransactionRepository) FindByAccountIDAndPeriod(accountID string, startDate, endDate string) ([]*models.Transaction, error) {
	query := `SELECT id, from_account_id, to_account_id, amount, type, status, description, created_at 
			  FROM transactions 
			  WHERE (from_account_id = $1 OR to_account_id = $1) 
			  AND created_at BETWEEN $2 AND $3
			  ORDER BY created_at DESC`

	rows, err := r.db.Query(query, accountID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения транзакций за период: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount,
			&tx.Type, &tx.Status, &tx.Description, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// Создание транзакции с транзакцией
func (r *TransactionRepository) CreateWithTx(tx *sql.Tx, transaction *models.Transaction) error {
	query := `INSERT INTO transactions (id, from_account_id, to_account_id, amount, type, status, description, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := tx.Exec(query, transaction.ID, transaction.FromAccountID, transaction.ToAccountID,
		transaction.Amount, transaction.Type, transaction.Status,
		transaction.Description, transaction.CreatedAt)

	return err
}

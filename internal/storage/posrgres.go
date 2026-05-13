package storage

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// ========== USER METHODS ==========

func (s *PostgresStorage) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at) 
              VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)
	return err
}

func (s *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at 
              FROM users WHERE email = $1`

	user := &models.User{}
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (s *PostgresStorage) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at 
              FROM users WHERE id = $1`

	user := &models.User{}
	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (s *PostgresStorage) GetUserEmail(userID string) string {
	query := `SELECT email FROM users WHERE id = $1`
	var email string
	err := s.db.QueryRow(query, userID).Scan(&email)
	if err != nil {
		return ""
	}
	return email
}

// ========== ACCOUNT METHODS ==========

func (s *PostgresStorage) CreateAccount(account *models.Account) error {
	query := `INSERT INTO accounts (id, user_id, account_number, balance, currency, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := s.db.Exec(query, account.ID, account.UserID, account.AccountNumber,
		account.Balance, account.Currency, account.CreatedAt)
	return err
}

func (s *PostgresStorage) GetAccountByID(id string) (*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
              FROM accounts WHERE id = $1`

	account := &models.Account{}
	err := s.db.QueryRow(query, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber,
		&account.Balance, &account.Currency, &account.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	return account, err
}

func (s *PostgresStorage) GetAccountsByUserID(userID string) ([]*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
              FROM accounts WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
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

func (s *PostgresStorage) GetAccountByNumber(number string) (*models.Account, error) {
	query := `SELECT id, user_id, account_number, balance, currency, created_at 
              FROM accounts WHERE account_number = $1`

	account := &models.Account{}
	err := s.db.QueryRow(query, number).Scan(
		&account.ID, &account.UserID, &account.AccountNumber,
		&account.Balance, &account.Currency, &account.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	return account, err
}

func (s *PostgresStorage) UpdateAccount(account *models.Account) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := s.db.Exec(query, account.Balance, account.ID)
	return err
}

// ========== TRANSACTION METHODS ==========

func (s *PostgresStorage) CreateTransaction(transaction *models.Transaction) error {
	query := `INSERT INTO transactions (id, from_account_id, to_account_id, amount, type, status, description, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(query, transaction.ID, transaction.FromAccountID, transaction.ToAccountID,
		transaction.Amount, transaction.Type, transaction.Status,
		transaction.Description, transaction.CreatedAt)
	return err
}

func (s *PostgresStorage) GetTransactionsByAccountID(accountID string) ([]*models.Transaction, error) {
	query := `SELECT id, from_account_id, to_account_id, amount, type, status, description, created_at 
              FROM transactions 
              WHERE from_account_id = $1 OR to_account_id = $1 
              ORDER BY created_at DESC
              LIMIT 100`

	rows, err := s.db.Query(query, accountID)
	if err != nil {
		return nil, err
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

// ========== CARD METHODS ==========

func (s *PostgresStorage) CreateCard(card *models.Card) error {
	query := `INSERT INTO cards (id, account_id, card_number_encrypted, card_number_masked, 
              expiry_encrypted, expiry_masked, cvv_hash, hmac_signature, is_active, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.Exec(query, card.ID, card.AccountID, card.CardNumberEncrypted, card.CardNumberMasked,
		card.ExpiryEncrypted, card.ExpiryMasked, card.CVVHash,
		card.HMACSignature, card.IsActive, card.CreatedAt)
	return err
}

func (s *PostgresStorage) GetCardsByAccountID(accountID string) ([]*models.Card, error) {
	query := `SELECT id, account_id, card_number_masked, expiry_masked, is_active, created_at 
              FROM cards WHERE account_id = $1 AND is_active = true`

	rows, err := s.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
		err := rows.Scan(&card.ID, &card.AccountID, &card.CardNumberMasked,
			&card.ExpiryMasked, &card.IsActive, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

// ========== CREDIT METHODS ==========

func (s *PostgresStorage) CreateCredit(credit *models.Credit) error {
	query := `INSERT INTO credits (id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.Exec(query, credit.ID, credit.UserID, credit.AccountID, credit.Amount,
		credit.InterestRate, credit.TermMonths, credit.MonthlyPayment,
		credit.RemainingAmount, credit.Status, credit.IssuedAt)
	return err
}

func (s *PostgresStorage) GetCreditByID(id string) (*models.Credit, error) {
	query := `SELECT id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at 
              FROM credits WHERE id = $1`

	credit := &models.Credit{}
	err := s.db.QueryRow(query, id).Scan(
		&credit.ID, &credit.UserID, &credit.AccountID, &credit.Amount,
		&credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment,
		&credit.RemainingAmount, &credit.Status, &credit.IssuedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("credit not found")
	}
	return credit, err
}

func (s *PostgresStorage) GetCreditsByUserID(userID string) ([]*models.Credit, error) {
	query := `SELECT id, user_id, account_id, amount, interest_rate, term_months, 
              monthly_payment, remaining_amount, status, issued_at 
              FROM credits WHERE user_id = $1 ORDER BY issued_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
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

func (s *PostgresStorage) UpdateCredit(credit *models.Credit) error {
	query := `UPDATE credits SET remaining_amount = $1, status = $2 WHERE id = $3`
	_, err := s.db.Exec(query, credit.RemainingAmount, credit.Status, credit.ID)
	return err
}

// ========== PAYMENT SCHEDULE METHODS ==========

func (s *PostgresStorage) CreatePaymentSchedule(creditID string, schedules []*models.PaymentSchedule) error {
	// Используем транзакцию для массовой вставки
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

	return tx.Commit()
}

func (s *PostgresStorage) GetPaymentSchedules(creditID string) ([]*models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, payment_number, due_date, amount, principal_amount, 
              interest_amount, penalty_amount, status, paid_at 
              FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_number`

	rows, err := s.db.Query(query, creditID)
	if err != nil {
		return nil, err
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

func (s *PostgresStorage) UpdatePaymentSchedule(schedule *models.PaymentSchedule) error {
	query := `UPDATE payment_schedules SET status = $1, paid_at = $2, penalty_amount = $3 
              WHERE id = $4`
	_, err := s.db.Exec(query, schedule.Status, schedule.PaidAt, schedule.PenaltyAmount, schedule.ID)
	return err
}

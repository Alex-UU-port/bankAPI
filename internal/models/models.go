package models

import (
	"time"
)

// Пользователь банка
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Банковский счет
type Account struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

// Банковская карта
type Card struct {
	ID                  string    `json:"id"`
	AccountID           string    `json:"account_id"`
	CardNumberEncrypted string    `json:"-"`
	CardNumberMasked    string    `json:"card_number"`
	ExpiryEncrypted     string    `json:"-"`
	ExpiryMasked        string    `json:"expiry"`
	CVVHash             string    `json:"-"`
	HMACSignature       string    `json:"-"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

// Транзакция
type Transaction struct {
	ID            string    `json:"id"`
	FromAccountID *string   `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Кредит
type Credit struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	AccountID       string    `json:"account_id"`
	Amount          float64   `json:"amount"`
	InterestRate    float64   `json:"interest_rate"`
	TermMonths      int       `json:"term_months"`
	MonthlyPayment  float64   `json:"monthly_payment"`
	RemainingAmount float64   `json:"remaining_amount"`
	Status          string    `json:"status"`
	IssuedAt        time.Time `json:"issued_at"`
}

// График платежей
type PaymentSchedule struct {
	ID              string     `json:"id"`
	CreditID        string     `json:"credit_id"`
	PaymentNumber   int        `json:"payment_number"`
	DueDate         time.Time  `json:"due_date"`
	Amount          float64    `json:"amount"`
	PrincipalAmount float64    `json:"principal_amount"`
	InterestAmount  float64    `json:"interest_amount"`
	PenaltyAmount   float64    `json:"penalty_amount"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
}

// DTO для запросов
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

type TransferRequest struct {
	ToAccountNumber string  `json:"to_account_number"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
}

type CreateCardRequest struct {
	AccountID string `json:"account_id"`
}

type CreateCreditRequest struct {
	AccountID  string  `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
}

type AnalyticsResponse struct {
	TotalIncome  float64            `json:"total_income"`
	TotalExpense float64            `json:"total_expense"`
	Balance      float64            `json:"balance"`
	CreditLoad   float64            `json:"credit_load"`
	MonthlyStats map[string]float64 `json:"monthly_stats"`
	KeyRate      float64            `json:"key_rate"`
}

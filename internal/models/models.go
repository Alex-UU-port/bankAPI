// internal/models/models.go
package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Account struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

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

type Transaction struct {
	ID            string    `json:"id"`
	FromAccountID string    `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccountNumber string  `json:"to_account_number"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
}

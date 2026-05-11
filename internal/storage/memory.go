// internal/storage/memory.go
package storage

import (
	"bankAPI/internal/models"
	"fmt"
	"sync"
)

type MemoryStorage struct {
	mu           sync.RWMutex
	users        map[string]*models.User
	accounts     map[string]*models.Account
	cards        map[string]*models.Card
	transactions map[string]*models.Transaction
	emailByUser  map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:        make(map[string]*models.User),
		accounts:     make(map[string]*models.Account),
		cards:        make(map[string]*models.Card),
		transactions: make(map[string]*models.Transaction),
		emailByUser:  make(map[string]string),
	}
}

// User methods
func (s *MemoryStorage) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return fmt.Errorf("user already exists")
	}
	s.users[user.ID] = user
	s.emailByUser[user.ID] = user.Email
	return nil
}

func (s *MemoryStorage) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// Account methods
func (s *MemoryStorage) CreateAccount(account *models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts[account.ID] = account
	return nil
}

func (s *MemoryStorage) GetAccountByID(id string) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, exists := s.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found")
	}
	return account, nil
}

func (s *MemoryStorage) GetAccountsByUserID(userID string) []*models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var accounts []*models.Account
	for _, acc := range s.accounts {
		if acc.UserID == userID {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

func (s *MemoryStorage) UpdateAccount(account *models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[account.ID]; !exists {
		return fmt.Errorf("account not found")
	}
	s.accounts[account.ID] = account
	return nil
}

// Card methods
func (s *MemoryStorage) CreateCard(card *models.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cards[card.ID] = card
	return nil
}

func (s *MemoryStorage) GetCardsByAccountID(accountID string) []*models.Card {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cards []*models.Card
	for _, card := range s.cards {
		if card.AccountID == accountID {
			cards = append(cards, card)
		}
	}
	return cards
}

// Transaction methods
func (s *MemoryStorage) CreateTransaction(tx *models.Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transactions[tx.ID] = tx
	return nil
}

func (s *MemoryStorage) GetUserEmail(userID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.emailByUser[userID]
}

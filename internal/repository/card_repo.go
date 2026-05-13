package repository

import (
	"bankAPI/internal/models"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// Создание карты
func (r *CardRepository) Create(card *models.Card) error {
	query := `INSERT INTO cards (id, account_id, card_number_encrypted, card_number_masked, 
              expiry_encrypted, expiry_masked, cvv_hash, hmac_signature, is_active, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(query, card.ID, card.AccountID, card.CardNumberEncrypted, card.CardNumberMasked,
		card.ExpiryEncrypted, card.ExpiryMasked, card.CVVHash, card.HMACSignature, card.IsActive, card.CreatedAt)

	if err != nil {
		logrus.WithError(err).Error("Ошибка создания карты")
		return fmt.Errorf("ошибка создания карты: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"card_id":    card.ID,
		"account_id": card.AccountID,
		"masked":     card.CardNumberMasked,
	}).Info("Карта успешно создана")

	return nil
}

// Поиск карты по ID
func (r *CardRepository) FindByID(id string) (*models.Card, error) {
	query := `SELECT id, account_id, card_number_encrypted, card_number_masked, 
              expiry_encrypted, expiry_masked, cvv_hash, hmac_signature, is_active, created_at 
              FROM cards WHERE id = $1`

	card := &models.Card{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.AccountID, &card.CardNumberEncrypted, &card.CardNumberMasked,
		&card.ExpiryEncrypted, &card.ExpiryMasked, &card.CVVHash, &card.HMACSignature,
		&card.IsActive, &card.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска карты: %w", err)
	}

	return card, nil
}

// Поиск карт по счету
func (r *CardRepository) FindByAccountID(accountID string) ([]*models.Card, error) {
	query := `SELECT id, account_id, card_number_masked, expiry_masked, is_active, created_at 
              FROM cards WHERE account_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения карт: %w", err)
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

// Обновление статуса карты
func (r *CardRepository) UpdateStatus(cardID string, isActive bool) error {
	query := `UPDATE cards SET is_active = $1 WHERE id = $2`
	_, err := r.db.Exec(query, isActive, cardID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса карты: %w", err)
	}
	return nil
}

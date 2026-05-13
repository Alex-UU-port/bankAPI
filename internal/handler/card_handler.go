package handler

import (
	"bankAPI/internal/models"
	"bankAPI/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{
		cardService: cardService,
	}
}

// Выпуск карты
func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	card, cvv, err := h.cardService.CreateCard(req.AccountID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card_id":     card.ID,
		"card_number": card.CardNumberMasked,
		"expiry":      card.ExpiryMasked,
		"cvv":         cvv,
		"message":     "Карта успешно выпущена, сохраните CVV код",
	})
}

// Получение карт по счету
func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	cards, err := h.cardService.GetUserCards(accountID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// Оплата картой
func (h *CardHandler) Pay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardID := vars["id"]

	var req struct {
		CVV    string  `json:"cvv"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.cardService.ProcessPayment(cardID, req.CVV, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Платеж успешно выполнен",
	})
}

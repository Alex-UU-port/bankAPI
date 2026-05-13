package handler

import (
	"bankAPI/internal/service"
	"encoding/json"
	"net/http"
)

type TransferHandler struct {
	transferService *service.TransferService
}

func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

// Перевод средств
func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req struct {
		FromAccountID   string  `json:"from_account_id"`
		ToAccountNumber string  `json:"to_account_number"`
		Amount          float64 `json:"amount"`
		Description     string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.transferService.Transfer(userID, req.FromAccountID, req.ToAccountNumber, req.Amount, req.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Перевод успешно выполнен",
	})
}

package handler

import (
	"bankAPI/internal/models"
	"bankAPI/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// Создание счета
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.CreateAccountRequest
	json.NewDecoder(r.Body).Decode(&req)

	account, err := h.accountService.CreateAccount(userID, req.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// Получение всех счетов пользователя
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	accounts, err := h.accountService.GetUserAccounts(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// Пополнение счета
func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	accountID := vars["id"]

	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.accountService.Deposit(accountID, userID, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Счет успешно пополнен"})
}

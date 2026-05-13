package handler

import (
	"bankAPI/internal/models"
	"bankAPI/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type CreditHandler struct {
	creditService *service.CreditService
}

func NewCreditHandler(creditService *service.CreditService) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
	}
}

// Оформление кредита
func (h *CreditHandler) CreateCredit(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.CreateCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	credit, schedules, err := h.creditService.CreateCredit(userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credit_id":       credit.ID,
		"amount":          credit.Amount,
		"interest_rate":   credit.InterestRate,
		"monthly_payment": credit.MonthlyPayment,
		"term_months":     credit.TermMonths,
		"schedules":       schedules,
		"message":         "Кредит успешно оформлен",
	})
}

// Получение графика платежей
func (h *CreditHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	creditID := vars["id"]

	schedules, err := h.creditService.GetPaymentSchedule(creditID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

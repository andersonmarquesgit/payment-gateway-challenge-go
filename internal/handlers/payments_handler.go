package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaymentsHandler struct {
	storage    *repository.PaymentsRepository
	bankClient *bank.Client
}

func NewPaymentsHandler(storage *repository.PaymentsRepository, bankClient *bank.Client) *PaymentsHandler {
	return &PaymentsHandler{
		storage:    storage,
		bankClient: bankClient,
	}
}

// GetHandler retrieves a payment by ID
// @Summary Get payment
// @Description Retrieves a previously processed payment
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.PostPaymentResponse
// @Failure 204 {string} string
// @Router /api/payments/{id} [get]
func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := h.storage.GetPayment(id)

		if payment != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(payment); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// PostHandler handles payment creation
// @Summary Create a payment
// @Description Processes a card payment through the payment gateway
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body models.PostPaymentRequest true "Payment request"
// @Success 200 {object} models.PostPaymentResponse
// @Failure 400 {object} models.PostPaymentResponse
// @Failure 502 {string} string
// @Router /api/payments [post]
func (ph *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req models.PostPaymentRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate Request
		if err := validation.ValidatePaymentRequest(req); err != nil {

			rejected := models.PostPaymentResponse{
				Id:                 uuid.NewString(),
				PaymentStatus:      models.Rejected,
				CardNumberLastFour: lastFour(req.CardNumber),
				ExpiryMonth:        req.ExpiryMonth,
				ExpiryYear:         req.ExpiryYear,
				Currency:           req.Currency,
				Amount:             req.Amount,
			}

			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(rejected)
			return
		}

		// Call bank simulator
		bankReq := bank.BankRequest{
			CardNumber: req.CardNumber,
			ExpiryDate: fmt.Sprintf("%02d/%d", req.ExpiryMonth, req.ExpiryYear),
			Currency:   req.Currency,
			Amount:     req.Amount,
			CVV:        req.Cvv,
		}

		bankResp, statusCode, err := ph.bankClient.ProcessPayment(bankReq)

		if err != nil {
			http.Error(w, "bank error", http.StatusBadGateway)
			return
		}

		if statusCode == http.StatusServiceUnavailable {
			http.Error(w, "acquiring bank unavailable", http.StatusBadGateway)
			return
		}

		// Determine status
		status := models.Declined
		if bankResp.Authorized {
			status = models.Authorized
		}

		payment := models.PostPaymentResponse{
			Id:                 uuid.NewString(),
			PaymentStatus:      status,
			CardNumberLastFour: lastFour(req.CardNumber),
			ExpiryMonth:        req.ExpiryMonth,
			ExpiryYear:         req.ExpiryYear,
			Currency:           req.Currency,
			Amount:             req.Amount,
		}

		// Persist
		ph.storage.AddPayment(payment)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payment)
	}
}

func lastFour(card string) string {
	if len(card) < 4 {
		return card
	}
	return card[len(card)-4:]
}

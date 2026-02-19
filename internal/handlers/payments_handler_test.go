package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client/bank"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaymentHandler(t *testing.T) {
	payment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: "4111111111111111",
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := repository.NewPaymentsRepository()
	ps.AddPayment(payment)

	payments := NewPaymentsHandler(ps, bank.NewClient())

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", payments.GetHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	t.Run("PaymentFound", func(t *testing.T) {
		// Create a new HTTP request for testing
		req, _ := http.NewRequest("GET", "/api/payments/test-id", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the body is not nil
		assert.NotNil(t, w.Body)

		// Check the HTTP status code in the response
		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
	t.Run("PaymentNotFound", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, _ := http.NewRequest("GET", "/api/payments/NonExistingID", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestPostPayment_Authorized(t *testing.T) {
	repo := repository.NewPaymentsRepository()
	handler := NewPaymentsHandler(repo, bank.NewClient())

	server := httptest.NewServer(handler.PostHandler())
	defer server.Close()

	payload := `{
		"card_number": "4111111111111111",
		"expiry_month": 12,
		"expiry_year": 2026,
		"currency": "GBP",
		"amount": 1000,
		"cvv": "123"
	}`

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(payload))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.PostPaymentResponse
	json.NewDecoder(resp.Body).Decode(&response)

	require.Equal(t, models.Authorized, response.PaymentStatus)
}

func TestPostPayment_Declined(t *testing.T) {
	repo := repository.NewPaymentsRepository()
	handler := NewPaymentsHandler(repo, bank.NewClient())

	server := httptest.NewServer(handler.PostHandler())
	defer server.Close()

	payload := `{
		"card_number": "4111111111111112",
		"expiry_month": 12,
		"expiry_year": 2026,
		"currency": "GBP",
		"amount": 1000,
		"cvv": "123"
	}`

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(payload))
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.PostPaymentResponse
	json.NewDecoder(resp.Body).Decode(&response)

	require.Equal(t, models.Declined, response.PaymentStatus)
}

func TestPostPayment_Rejected(t *testing.T) {
	repo := repository.NewPaymentsRepository()
	handler := NewPaymentsHandler(repo, bank.NewClient())

	server := httptest.NewServer(handler.PostHandler())
	defer server.Close()

	payload := `{
		"card_number": "4111111111111111",
		"expiry_month": 12,
		"expiry_year": 2026,
		"currency": "GBP",
		"amount": 1000,
		"cvv": "12"
	}`

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(payload))
	require.NoError(t, err)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response models.PostPaymentResponse
	json.NewDecoder(resp.Body).Decode(&response)

	require.Equal(t, models.Rejected, response.PaymentStatus)
}

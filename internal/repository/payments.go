package repository

import (
	"sync"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type PaymentsRepository struct {
	mu       sync.RWMutex
	payments []models.PostPaymentResponse
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{
		payments: []models.PostPaymentResponse{},
	}
}

func (ps *PaymentsRepository) GetPayment(id string) *models.PostPaymentResponse {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, element := range ps.payments {
		if element.Id == id {
			return &element
		}
	}
	return nil
}

func (ps *PaymentsRepository) AddPayment(payment models.PostPaymentResponse) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	ps.payments = append(ps.payments, payment)
}

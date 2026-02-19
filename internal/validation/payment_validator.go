package validation

import (
	"errors"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

var allowedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
}

func ValidatePaymentRequest(req models.PostPaymentRequest) error {

	if len(req.CardNumber) < 14 || len(req.CardNumber) > 19 {
		return errors.New("invalid card number length")
	}

	if !isNumeric(req.CardNumber) {
		return errors.New("card number must be numeric")
	}

	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return errors.New("invalid expiry month")
	}

	if !isFutureDate(req.ExpiryMonth, req.ExpiryYear) {
		return errors.New("card expired")
	}

	if !allowedCurrencies[req.Currency] {
		return errors.New("unsupported currency")
	}

	if req.Amount <= 0 {
		return errors.New("invalid amount")
	}

	if len(req.Cvv) < 3 || len(req.Cvv) > 4 || !isNumeric(req.Cvv) {
		return errors.New("invalid cvv")
	}

	return nil
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isFutureDate(month, year int) bool {
	now := time.Now()
	currentYear, currentMonth := now.Year(), int(now.Month())

	if year < currentYear {
		return false
	}

	if year == currentYear && month < currentMonth {
		return false
	}

	return true
}

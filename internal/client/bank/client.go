package bank

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type BankRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	CVV        string `json:"cvv"`
}

type BankResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}

func NewClient() *Client {
	bankURL := os.Getenv("BANK_SIMULATOR_URL")
	if bankURL == "" {
		bankURL = "http://localhost:8080"
	}

	return &Client{
		baseURL: bankURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) ProcessPayment(req BankRequest) (*BankResponse, int, error) {
	body, _ := json.Marshal(req)

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/payments", c.baseURL),
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}

	var bankResp BankResponse
	if err := json.NewDecoder(resp.Body).Decode(&bankResp); err != nil {
		return nil, 0, err
	}

	return &bankResp, resp.StatusCode, nil
}

# Payment Gateway Challenge

## Overview

This project implements a simple Payment Gateway API that allows merchants to:

- Process card payments through a simulated acquiring bank
- Retrieve previously processed payments

The gateway validates incoming payment requests, forwards valid requests to a bank simulator, stores payment results in-memory, and returns an appropriate status to the merchant.

The implementation focuses on simplicity, correctness, and maintainability while respecting the constraints defined in the challenge.

---

## Architecture

The application follows a layered structure to keep responsibilities clearly separated:

- **API / Handler layer**
    - Handles HTTP concerns such as request parsing, response formatting, and status codes.

- **Validation layer**
    - Performs all request validations before any interaction with the acquiring bank.

- **Bank client**
    - Encapsulates communication with the acquiring bank simulator.

- **Repository layer**
    - Provides thread-safe, in-memory storage for processed payments.

- **Domain models**
    - Represent payment requests, responses, and statuses.

This separation allows the system to be easily testable, maintainable, and extensible without unnecessary complexity.

---

## Design Decisions

- **Rejected vs Declined**
    - *Rejected* payments are returned when the request fails validation and the acquiring bank is not called.
    - *Declined* payments represent valid requests that were sent to the acquiring bank but were not authorized.

- **Sensitive data handling**
    - Full card numbers (PAN) and CVV values are never stored or returned.
    - Only the last four digits of the card number are exposed in responses.

- **Payment identifiers**
    - UUIDs are used to uniquely identify payments due to their simplicity and low collision probability.

- **In-memory storage**
    - An in-memory repository is used as required by the challenge, with proper concurrency protection.

- **Bank failures**
    - If the acquiring bank is unavailable (503), the payment is not persisted and the gateway returns a 502 response.

---

## Assumptions

- The acquiring bank behavior is deterministic and defined by the simulator:
    - Card numbers ending with an odd number (1, 3, 5, 7, 9): Authorized
    - Card numbers ending with an even number (2, 4, 6, 8): Declined
    - Card numbers ending zero (0): Bank unavailable (503 Service Unavailable)

- Currency validation is limited to a small, predefined subset as required by the challenge.

- Persistence is not required beyond the lifecycle of the application.

---

## Improvements for Production

If this gateway were to be production-ready, the following improvements would be considered:

- Persistent storage (e.g. relational database)
- Idempotency keys to avoid duplicate payments
- Retry mechanisms with exponential backoff
- Circuit breaker for bank communication
- Structured logging and metrics
- Distributed tracing
- Tokenization or encryption of sensitive card data
- Horizontal scalability and stateless deployments

---

## How to Run

1. Start the bank simulator:
```bash
docker-compose up
```

2. Start the payment gateway API:
```bash
go run main.go
```

## How to Test
The API can be tested using Postman or via automated tests.

Create a Payment
```
curl --location 'http://localhost:8090/api/payments' \
--header 'Content-Type: application/json' \
--data '{
  "card_number": "4111111111111111",
  "expiry_month": 12,
  "expiry_year": 2026,
  "currency": "GBP",
  "amount": 1050,
  "cvv": "123"
}
'
```

POST /api/payments

Example payload:
```json
{
  "card_number": "4111111111111111",
  "expiry_month": 12,
  "expiry_year": 2026,
  "currency": "GBP",
  "amount": 1050,
  "cvv": "123"
}
```

## Template structure
```
main.go - a skeleton Payment Gateway API
imposters/ - contains the bank simulator configuration. Don't change this
docs/docs.go - Generated file by Swaggo
.editorconfig - don't change this. It ensures a consistent set of rules for submissions when reformatting code
docker-compose.yml - configures the bank simulator
.goreleaser.yml - Goreleaser configuration
```

Feel free to change the structure of the solution, use a different test library etc.

### Swagger
This template uses Swaggo to autodocument the API and create a Swagger spec. The Swagger UI is available at http://localhost:8090/swagger/index.html.

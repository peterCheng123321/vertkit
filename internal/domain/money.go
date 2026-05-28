package domain

import (
	"errors"
	"fmt"
	"math"
)

// Currency represents an ISO 4217 currency code.
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencyCNY Currency = "CNY"
	CurrencyINR Currency = "INR"
	CurrencyBRL Currency = "BRL"
	CurrencyAUD Currency = "AUD"
	CurrencyCAD Currency = "CAD"
	CurrencyCHF Currency = "CHF"
)

// Money represents a monetary value with its currency.
// It stores the amount in the currency's minor units (e.g. cents for USD)
// to avoid floating point errors. This is a core global primitive.
type Money struct {
	Amount   int64    `json:"amount"`
	Currency Currency `json:"currency"`
}

// NewMoney creates a Money value after basic validation.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount must be non-negative")
	}
	if !isValidCurrency(currency) {
		return Money{}, fmt.Errorf("unsupported currency: %s", currency)
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// MustNewMoney is like NewMoney but panics on error. Use only in tests or constants.
func MustNewMoney(amount int64, currency Currency) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

func isValidCurrency(c Currency) bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyJPY, CurrencyCNY,
		CurrencyINR, CurrencyBRL, CurrencyAUD, CurrencyCAD, CurrencyCHF:
		return true
	default:
		return false
	}
}

// Add adds two Money values. Currencies must match.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, errors.New("cannot add different currencies")
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Multiply scales the amount by a positive integer multiplier.
func (m Money) Multiply(multiplier int64) (Money, error) {
	if multiplier < 0 {
		return Money{}, errors.New("multiplier must be non-negative")
	}
	return Money{Amount: m.Amount * multiplier, Currency: m.Currency}, nil
}

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// String returns a simple representation, e.g. "USD 12345".
func (m Money) String() string {
	return fmt.Sprintf("%s %d", m.Currency, m.Amount)
}

// Format returns a human-friendly string. For production use a proper formatter
// that respects locale and currency minor units (e.g. 2 decimals for USD, 0 for JPY).
func (m Money) Format() string {
	// Minimal implementation. Real global apps should use golang.org/x/text/currency + message.
	factor := MinorUnitsPerUnit(m.Currency)
	if factor == 1 {
		return fmt.Sprintf("%s %d", m.Currency, m.Amount)
	}
	unit := float64(m.Amount) / float64(factor)
	return fmt.Sprintf("%s %.2f", m.Currency, unit)
}

// MinorUnitsPerUnit returns how many minor units make one major unit for the currency.
func MinorUnitsPerUnit(c Currency) int64 {
	switch c {
	case CurrencyJPY:
		return 1
	default:
		return 100
	}
}

// FromMajor creates Money from a major unit amount (e.g. 12.50 USD -> 1250).
func FromMajor(major float64, currency Currency) (Money, error) {
	if major < 0 {
		return Money{}, errors.New("amount must be non-negative")
	}
	if !isValidCurrency(currency) {
		return Money{}, fmt.Errorf("unsupported currency: %s", currency)
	}
	factor := float64(MinorUnitsPerUnit(currency))
	amount := int64(math.Round(major * factor))
	return NewMoney(amount, currency)
}

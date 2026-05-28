package domain

import (
	"errors"
	"strings"
)

// CountryCode is an ISO 3166-1 alpha-2 country code.
type CountryCode string

const (
	CountryUS CountryCode = "US"
	CountryGB CountryCode = "GB"
	CountryDE CountryCode = "DE"
	CountryFR CountryCode = "FR"
	CountryCN CountryCode = "CN"
	CountryIN CountryCode = "IN"
	CountryJP CountryCode = "JP"
	CountryBR CountryCode = "BR"
	CountryCA CountryCode = "CA"
	CountryAU CountryCode = "AU"
)

// Address represents a physical or postal address.
// The structure is intentionally global-first: every field that varies by jurisdiction
// is captured explicitly so that validation, formatting, and tax logic can be added
// without changing the core type later.
//
// For full correctness, real implementations will need per-country rules
// (e.g. Brazilian addresses have different fields than Japanese ones).
// This type provides the common substrate.
type Address struct {
	Line1       string      `json:"line1"`
	Line2       string      `json:"line2,omitempty"`
	City        string      `json:"city"`
	State       string      `json:"state,omitempty"` // Province, region, state, etc.
	PostalCode  string      `json:"postal_code,omitempty"`
	CountryCode CountryCode `json:"country_code"`
}

// NewAddress performs basic structural validation.
func NewAddress(line1, city string, country CountryCode) (Address, error) {
	if strings.TrimSpace(line1) == "" {
		return Address{}, errors.New("line1 is required")
	}
	if strings.TrimSpace(city) == "" {
		return Address{}, errors.New("city is required")
	}
	if !isValidCountry(country) {
		return Address{}, errors.New("country_code must be a valid ISO 3166-1 alpha-2 code")
	}
	return Address{
		Line1:       strings.TrimSpace(line1),
		City:        strings.TrimSpace(city),
		CountryCode: country,
	}, nil
}

func isValidCountry(c CountryCode) bool {
	switch c {
	case CountryUS, CountryGB, CountryDE, CountryFR, CountryCN,
		CountryIN, CountryJP, CountryBR, CountryCA, CountryAU:
		return true
	default:
		// In a real global system we would accept any valid ISO code.
		// For v0 we keep a small curated list to force explicit decisions.
		return len(c) == 2
	}
}

// FullCountryName returns a human name for the small set of supported countries.
func (a Address) FullCountryName() string {
	switch a.CountryCode {
	case CountryUS:
		return "United States"
	case CountryGB:
		return "United Kingdom"
	case CountryDE:
		return "Germany"
	case CountryFR:
		return "France"
	case CountryCN:
		return "China"
	case CountryIN:
		return "India"
	case CountryJP:
		return "Japan"
	case CountryBR:
		return "Brazil"
	case CountryCA:
		return "Canada"
	case CountryAU:
		return "Australia"
	default:
		return string(a.CountryCode)
	}
}

// IsEmpty reports whether the address has no meaningful content.
func (a Address) IsEmpty() bool {
	return a.Line1 == "" && a.City == "" && a.CountryCode == ""
}

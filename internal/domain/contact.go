package domain

import (
	"errors"
	"strings"
	"time"
)

// ContactID is the unique identifier for a person/contact.
type ContactID string

// Contact represents an individual person. In CRM this is usually linked to an Account,
// but can exist standalone (e.g. for B2C or early leads).
//
// Global: Title, phone numbers, and address follow the same patterns as Account.
type Contact struct {
	ID             ContactID      `json:"id"`
	TenantID       TenantID       `json:"tenant_id"`
	AccountID      AccountID      `json:"account_id,omitempty"` // optional for B2C / leads
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Email          string         `json:"email,omitempty"`
	Phone          string         `json:"phone,omitempty"`
	JobTitle       string         `json:"job_title,omitempty"`
	Department     string         `json:"department,omitempty"`
	MailingAddress Address        `json:"mailing_address"`
	Status         string         `json:"status"` // "active", "unsubscribed", "archived"
	OwnerID        string         `json:"owner_id,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// NewContact creates a contact with basic validation.
func NewContact(id ContactID, tenantID TenantID, firstName, lastName string) (Contact, error) {
	if id == "" {
		return Contact{}, errors.New("contact id is required")
	}
	if tenantID == "" {
		return Contact{}, errors.New("tenant_id is required")
	}
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	if firstName == "" || lastName == "" {
		return Contact{}, errors.New("first_name and last_name are required")
	}

	now := time.Now().UTC()
	return Contact{
		ID:        id,
		TenantID:  tenantID,
		FirstName: firstName,
		LastName:  lastName,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// FullName returns the combined name.
func (c Contact) FullName() string {
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

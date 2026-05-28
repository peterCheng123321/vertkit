package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vertkit/vertkit/internal/domain"
	"github.com/vertkit/vertkit/internal/rules"
	"github.com/vertkit/vertkit/internal/storage"
)

// Server holds dependencies for the HTTP API.
type Server struct {
	stores       *storage.Stores
	mux          *http.ServeMux
	serviceToken string
}

// Option configures the API server.
type Option func(*Server)

// WithServiceToken requires a bearer token for tenant-scoped routes.
func WithServiceToken(token string) Option {
	return func(s *Server) {
		s.serviceToken = token
	}
}

// NewServer creates the API server with routes.
func NewServer(stores *storage.Stores, opts ...Option) *Server {
	s := &Server{
		stores: stores,
		mux:    http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.registerRoutes()
	return s
}

func (s *Server) Router() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	// Tenants (not tenant-scoped)
	s.mux.HandleFunc("POST /tenants", s.handleCreateTenant)
	s.mux.HandleFunc("GET /tenants", s.handleListTenants)
	s.mux.HandleFunc("GET /tenants/{id}", s.handleGetTenant)
	s.mux.HandleFunc("PUT /tenants/{id}", s.handleUpdateTenant)
	s.mux.HandleFunc("DELETE /tenants/{id}", s.handleDeleteTenant)

	// Accounts
	s.mux.HandleFunc("POST /tenants/{tenant_id}/accounts", s.tenantScoped(s.handleCreateAccount))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/accounts", s.tenantScoped(s.handleListAccounts))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/accounts/{id}", s.tenantScoped(s.handleGetAccount))
	s.mux.HandleFunc("PUT /tenants/{tenant_id}/accounts/{id}", s.tenantScoped(s.handleUpdateAccount))
	s.mux.HandleFunc("DELETE /tenants/{tenant_id}/accounts/{id}", s.tenantScoped(s.handleDeleteAccount))

	// Contacts
	s.mux.HandleFunc("POST /tenants/{tenant_id}/contacts", s.tenantScoped(s.handleCreateContact))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/contacts", s.tenantScoped(s.handleListContacts))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/contacts/{id}", s.tenantScoped(s.handleGetContact))
	s.mux.HandleFunc("PUT /tenants/{tenant_id}/contacts/{id}", s.tenantScoped(s.handleUpdateContact))
	s.mux.HandleFunc("DELETE /tenants/{tenant_id}/contacts/{id}", s.tenantScoped(s.handleDeleteContact))

	// Opportunities
	s.mux.HandleFunc("POST /tenants/{tenant_id}/opportunities", s.tenantScoped(s.handleCreateOpportunity))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/opportunities", s.tenantScoped(s.handleListOpportunities))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/opportunities/{id}", s.tenantScoped(s.handleGetOpportunity))
	s.mux.HandleFunc("PUT /tenants/{tenant_id}/opportunities/{id}", s.tenantScoped(s.handleUpdateOpportunity))
	s.mux.HandleFunc("DELETE /tenants/{tenant_id}/opportunities/{id}", s.tenantScoped(s.handleDeleteOpportunity))

	// Products
	s.mux.HandleFunc("POST /tenants/{tenant_id}/products", s.tenantScoped(s.handleCreateProduct))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/products", s.tenantScoped(s.handleListProducts))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/products/{id}", s.tenantScoped(s.handleGetProduct))
	s.mux.HandleFunc("PUT /tenants/{tenant_id}/products/{id}", s.tenantScoped(s.handleUpdateProduct))
	s.mux.HandleFunc("DELETE /tenants/{tenant_id}/products/{id}", s.tenantScoped(s.handleDeleteProduct))

	// Business Rules (the new slice)
	s.mux.HandleFunc("POST /tenants/{tenant_id}/rules", s.tenantScoped(s.handleCreateRule))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/rules", s.tenantScoped(s.handleListRules))
	s.mux.HandleFunc("GET /tenants/{tenant_id}/rules/{id}", s.tenantScoped(s.handleGetRule))
	s.mux.HandleFunc("PUT /tenants/{tenant_id}/rules/{id}", s.tenantScoped(s.handleUpdateRule))
	s.mux.HandleFunc("DELETE /tenants/{tenant_id}/rules/{id}", s.tenantScoped(s.handleDeleteRule))
	s.mux.HandleFunc("POST /tenants/{tenant_id}/rules/evaluate/opportunity", s.tenantScoped(s.handleEvaluateOpportunityRules))

	// Health
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func tenantIDFromPath(r *http.Request) (domain.TenantID, error) {
	id := r.PathValue("tenant_id")
	if id == "" {
		return "", errors.New("missing tenant_id in path")
	}
	return domain.TenantID(id), nil
}

func (s *Server) tenantScoped(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authorize(w, r) {
			return
		}
		tenantID, err := tenantIDFromPath(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if _, err := s.stores.Tenants.Get(r.Context(), tenantID); err != nil {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		next(w, r)
	}
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request) bool {
	if s.serviceToken == "" {
		return true
	}
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", s.serviceToken) {
		writeError(w, http.StatusUnauthorized, "missing or invalid service token")
		return false
	}
	return true
}

// --- Tenant Handlers ---

type createTenantReq struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DefaultCurrency string `json:"default_currency"`
	DefaultLocale   string `json:"default_locale"`
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	t, err := domain.NewTenant(
		domain.TenantID(req.ID),
		req.Name,
		domain.Currency(req.DefaultCurrency),
		req.DefaultLocale,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.stores.Tenants.Create(r.Context(), t); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	list, err := s.stores.Tenants.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id := domain.TenantID(r.PathValue("id"))
	t, err := s.stores.Tenants.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	id := domain.TenantID(r.PathValue("id"))
	existing, err := s.stores.Tenants.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	var req createTenantReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID == "" {
		req.ID = string(id)
	}
	if domain.TenantID(req.ID) != id {
		writeError(w, http.StatusBadRequest, "tenant id cannot be changed")
		return
	}
	t, err := domain.NewTenant(id, req.Name, domain.Currency(req.DefaultCurrency), req.DefaultLocale)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	t.CreatedAt = existing.CreatedAt
	t.UpdatedAt = time.Now().UTC()
	if err := s.stores.Tenants.Update(r.Context(), t); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := domain.TenantID(r.PathValue("id"))
	if _, err := s.stores.Tenants.Get(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if err := s.stores.Tenants.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Account Handlers ---

type createAccountReq struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Website      string         `json:"website,omitempty"`
	Industry     string         `json:"industry,omitempty"`
	Status       string         `json:"status,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req createAccountReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	acc, err := domain.NewAccount(domain.AccountID(req.ID), tenantID, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	acc.Website = req.Website
	acc.Industry = req.Industry
	if req.Status != "" {
		acc.Status = req.Status
	}
	acc.CustomFields = req.CustomFields

	if err := s.stores.Accounts.Create(r.Context(), tenantID, acc); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := s.stores.Accounts.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.AccountID(r.PathValue("id"))
	acc, err := s.stores.Accounts.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.AccountID(r.PathValue("id"))
	existing, err := s.stores.Accounts.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}

	var req createAccountReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID != "" && domain.AccountID(req.ID) != id {
		writeError(w, http.StatusBadRequest, "account id cannot be changed")
		return
	}
	acc, err := domain.NewAccount(id, tenantID, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	acc.Website = req.Website
	acc.Industry = req.Industry
	if req.Status != "" {
		acc.Status = req.Status
	} else {
		acc.Status = existing.Status
	}
	if req.CustomFields != nil {
		acc.CustomFields = req.CustomFields
	} else {
		acc.CustomFields = existing.CustomFields
	}
	acc.CreatedAt = existing.CreatedAt
	acc.UpdatedAt = time.Now().UTC()
	if err := s.stores.Accounts.Update(r.Context(), tenantID, acc); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.AccountID(r.PathValue("id"))
	if err := s.stores.Accounts.Delete(r.Context(), tenantID, id); err != nil {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Contact Handlers (minimal) ---

type createContactReq struct {
	ID           string         `json:"id"`
	AccountID    string         `json:"account_id,omitempty"`
	FirstName    string         `json:"first_name"`
	LastName     string         `json:"last_name"`
	Email        string         `json:"email,omitempty"`
	Phone        string         `json:"phone,omitempty"`
	JobTitle     string         `json:"job_title,omitempty"`
	Department   string         `json:"department,omitempty"`
	Status       string         `json:"status,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

func (s *Server) handleCreateContact(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req createContactReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.AccountID != "" {
		if _, err := s.stores.Accounts.Get(r.Context(), tenantID, domain.AccountID(req.AccountID)); err != nil {
			writeError(w, http.StatusBadRequest, "account not found")
			return
		}
	}

	ct, err := domain.NewContact(domain.ContactID(req.ID), tenantID, req.FirstName, req.LastName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ct.AccountID = domain.AccountID(req.AccountID)
	ct.Email = req.Email
	ct.Phone = req.Phone
	ct.JobTitle = req.JobTitle
	ct.Department = req.Department
	if req.Status != "" {
		ct.Status = req.Status
	}
	ct.CustomFields = req.CustomFields

	if err := s.stores.Contacts.Create(r.Context(), tenantID, ct); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ct)
}

func (s *Server) handleListContacts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := s.stores.Contacts.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ct, err := s.stores.Contacts.Get(r.Context(), tenantID, domain.ContactID(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, "contact not found")
		return
	}
	writeJSON(w, http.StatusOK, ct)
}

func (s *Server) handleUpdateContact(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.ContactID(r.PathValue("id"))
	existing, err := s.stores.Contacts.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "contact not found")
		return
	}
	var req createContactReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID != "" && domain.ContactID(req.ID) != id {
		writeError(w, http.StatusBadRequest, "contact id cannot be changed")
		return
	}
	if req.AccountID != "" {
		if _, err := s.stores.Accounts.Get(r.Context(), tenantID, domain.AccountID(req.AccountID)); err != nil {
			writeError(w, http.StatusBadRequest, "account not found")
			return
		}
	}
	ct, err := domain.NewContact(id, tenantID, req.FirstName, req.LastName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ct.AccountID = domain.AccountID(req.AccountID)
	ct.Email = req.Email
	ct.Phone = req.Phone
	ct.JobTitle = req.JobTitle
	ct.Department = req.Department
	if req.Status != "" {
		ct.Status = req.Status
	} else {
		ct.Status = existing.Status
	}
	if req.CustomFields != nil {
		ct.CustomFields = req.CustomFields
	} else {
		ct.CustomFields = existing.CustomFields
	}
	ct.CreatedAt = existing.CreatedAt
	ct.UpdatedAt = time.Now().UTC()
	if err := s.stores.Contacts.Update(r.Context(), tenantID, ct); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ct)
}

func (s *Server) handleDeleteContact(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.stores.Contacts.Delete(r.Context(), tenantID, domain.ContactID(r.PathValue("id"))); err != nil {
		writeError(w, http.StatusNotFound, "contact not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Opportunity Handlers (minimal) ---

type createOpportunityReq struct {
	ID           string         `json:"id"`
	AccountID    string         `json:"account_id"`
	ContactID    string         `json:"contact_id,omitempty"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Amount       int64          `json:"amount"` // minor units
	Currency     string         `json:"currency"`
	Stage        string         `json:"stage,omitempty"`
	Probability  int            `json:"probability,omitempty"`
	Source       string         `json:"source,omitempty"`
	OwnerID      string         `json:"owner_id,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

func (s *Server) handleCreateOpportunity(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req createOpportunityReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if _, err := s.stores.Accounts.Get(r.Context(), tenantID, domain.AccountID(req.AccountID)); err != nil {
		writeError(w, http.StatusBadRequest, "account not found")
		return
	}
	if req.ContactID != "" {
		if _, err := s.stores.Contacts.Get(r.Context(), tenantID, domain.ContactID(req.ContactID)); err != nil {
			writeError(w, http.StatusBadRequest, "contact not found")
			return
		}
	}

	money, err := domain.NewMoney(req.Amount, domain.Currency(req.Currency))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	opp, err := domain.NewOpportunity(
		domain.OpportunityID(req.ID),
		tenantID,
		domain.AccountID(req.AccountID),
		req.Name,
		money,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	opp.ContactID = domain.ContactID(req.ContactID)
	opp.Description = req.Description
	if req.Stage != "" {
		opp.Stage = req.Stage
	}
	if req.Probability != 0 {
		opp.Probability = req.Probability
	}
	opp.Source = req.Source
	opp.OwnerID = req.OwnerID
	opp.CustomFields = req.CustomFields

	if err := s.stores.Opportunities.Create(r.Context(), tenantID, opp); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, opp)
}

func (s *Server) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := s.stores.Opportunities.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	opp, err := s.stores.Opportunities.Get(r.Context(), tenantID, domain.OpportunityID(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}
	writeJSON(w, http.StatusOK, opp)
}

func (s *Server) handleUpdateOpportunity(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.OpportunityID(r.PathValue("id"))
	existing, err := s.stores.Opportunities.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}
	var req createOpportunityReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID != "" && domain.OpportunityID(req.ID) != id {
		writeError(w, http.StatusBadRequest, "opportunity id cannot be changed")
		return
	}
	if _, err := s.stores.Accounts.Get(r.Context(), tenantID, domain.AccountID(req.AccountID)); err != nil {
		writeError(w, http.StatusBadRequest, "account not found")
		return
	}
	if req.ContactID != "" {
		if _, err := s.stores.Contacts.Get(r.Context(), tenantID, domain.ContactID(req.ContactID)); err != nil {
			writeError(w, http.StatusBadRequest, "contact not found")
			return
		}
	}
	money, err := domain.NewMoney(req.Amount, domain.Currency(req.Currency))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	opp, err := domain.NewOpportunity(id, tenantID, domain.AccountID(req.AccountID), req.Name, money)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	opp.ContactID = domain.ContactID(req.ContactID)
	opp.Description = req.Description
	if req.Stage != "" {
		opp.Stage = req.Stage
	} else {
		opp.Stage = existing.Stage
	}
	if req.Probability != 0 {
		opp.Probability = req.Probability
	} else {
		opp.Probability = existing.Probability
	}
	opp.Source = req.Source
	opp.OwnerID = req.OwnerID
	if req.CustomFields != nil {
		opp.CustomFields = req.CustomFields
	} else {
		opp.CustomFields = existing.CustomFields
	}
	opp.CreatedAt = existing.CreatedAt
	opp.UpdatedAt = time.Now().UTC()
	if err := s.stores.Opportunities.Update(r.Context(), tenantID, opp); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, opp)
}

func (s *Server) handleDeleteOpportunity(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.stores.Opportunities.Delete(r.Context(), tenantID, domain.OpportunityID(r.PathValue("id"))); err != nil {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Product Handlers (minimal) ---

type createProductReq struct {
	ID                string         `json:"id"`
	SKU               string         `json:"sku,omitempty"`
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	Price             int64          `json:"price"` // minor units
	Currency          string         `json:"currency"`
	Unit              string         `json:"unit,omitempty"`
	IsRecurring       bool           `json:"is_recurring,omitempty"`
	RecurringInterval string         `json:"recurring_interval,omitempty"`
	Category          string         `json:"category,omitempty"`
	Active            *bool          `json:"active,omitempty"`
	CustomFields      map[string]any `json:"custom_fields,omitempty"`
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req createProductReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	money, err := domain.NewMoney(req.Price, domain.Currency(req.Currency))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	prod, err := domain.NewProduct(domain.ProductID(req.ID), tenantID, req.Name, money)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	prod.SKU = req.SKU
	prod.Description = req.Description
	prod.Unit = req.Unit
	prod.IsRecurring = req.IsRecurring
	prod.RecurringInterval = req.RecurringInterval
	prod.Category = req.Category
	if req.Active != nil {
		prod.Active = *req.Active
	}
	prod.CustomFields = req.CustomFields

	if err := s.stores.Products.Create(r.Context(), tenantID, prod); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, prod)
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := s.stores.Products.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	product, err := s.stores.Products.Get(r.Context(), tenantID, domain.ProductID(r.PathValue("id")))
	if err != nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (s *Server) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := domain.ProductID(r.PathValue("id"))
	existing, err := s.stores.Products.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	var req createProductReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID != "" && domain.ProductID(req.ID) != id {
		writeError(w, http.StatusBadRequest, "product id cannot be changed")
		return
	}
	money, err := domain.NewMoney(req.Price, domain.Currency(req.Currency))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	product, err := domain.NewProduct(id, tenantID, req.Name, money)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	product.SKU = req.SKU
	product.Description = req.Description
	product.Unit = req.Unit
	product.IsRecurring = req.IsRecurring
	product.RecurringInterval = req.RecurringInterval
	product.Category = req.Category
	if req.Active != nil {
		product.Active = *req.Active
	} else {
		product.Active = existing.Active
	}
	if req.CustomFields != nil {
		product.CustomFields = req.CustomFields
	} else {
		product.CustomFields = existing.CustomFields
	}
	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now().UTC()
	if err := s.stores.Products.Update(r.Context(), tenantID, product); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (s *Server) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.stores.Products.Delete(r.Context(), tenantID, domain.ProductID(r.PathValue("id"))); err != nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Shutdown is a no-op placeholder for graceful shutdown in the future.
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// --- Business Rules Handlers ---

type createRuleReq struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	EntityType  string                  `json:"entity_type"`
	Conditions  []storage.RuleCondition `json:"conditions"`
	Actions     []storage.RuleAction    `json:"actions"`
	IsActive    bool                    `json:"is_active"`
}

func (s *Server) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req createRuleReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID == "" || req.Name == "" || req.EntityType == "" {
		writeError(w, http.StatusBadRequest, "id, name, and entity_type are required")
		return
	}

	now := time.Now().UTC()
	rule := storage.Rule{
		ID:          req.ID,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		EntityType:  req.EntityType,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.stores.Rules.Create(r.Context(), tenantID, rule); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := s.stores.Rules.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetRule(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rule, err := s.stores.Rules.Get(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := r.PathValue("id")
	existing, err := s.stores.Rules.Get(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	var req createRuleReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.ID != "" && req.ID != id {
		writeError(w, http.StatusBadRequest, "rule id cannot be changed")
		return
	}
	if req.Name == "" || req.EntityType == "" {
		writeError(w, http.StatusBadRequest, "name and entity_type are required")
		return
	}
	rule := storage.Rule{
		ID:          id,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		EntityType:  req.EntityType,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		IsActive:    req.IsActive,
		CreatedAt:   existing.CreatedAt,
		UpdatedAt:   time.Now().UTC(),
	}
	if err := s.stores.Rules.Update(r.Context(), tenantID, rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.stores.Rules.Delete(r.Context(), tenantID, r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// evaluateOpportunityReq is the payload for running rules against an opportunity snapshot.
type evaluateOpportunityReq struct {
	Opportunity domain.Opportunity `json:"opportunity"`
	Operation   string             `json:"operation,omitempty"` // create, update, etc.
}

func (s *Server) handleEvaluateOpportunityRules(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req evaluateOpportunityReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	if req.Opportunity.TenantID != "" && req.Opportunity.TenantID != tenantID {
		writeError(w, http.StatusBadRequest, "opportunity tenant_id must match path tenant_id")
		return
	}
	req.Opportunity.TenantID = tenantID

	activeRules, err := s.stores.Rules.ListActive(r.Context(), tenantID, "opportunity")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert storage.Rule -> rules.Rule (thin adapter for this slice)
	engineRules := make([]rules.Rule, len(activeRules))
	for i, sr := range activeRules {
		engineRules[i] = rules.Rule{
			ID:          sr.ID,
			TenantID:    sr.TenantID,
			Name:        sr.Name,
			Description: sr.Description,
			EntityType:  sr.EntityType,
			Conditions:  convertConditions(sr.Conditions),
			Actions:     convertActions(sr.Actions),
			IsActive:    sr.IsActive,
		}
	}

	eng := rules.NewEngine()
	ctx := r.Context()
	ec := rules.EvaluationContext{
		Entity:     req.Opportunity,
		EntityType: "opportunity",
		TenantID:   tenantID,
		Operation:  req.Operation,
		Now:        time.Now().UTC(),
	}

	results := eng.Evaluate(ctx, engineRules, ec)

	// Summarize for the caller
	hasBlocking := false
	allErrors := []string{}
	allWarnings := []string{}
	for _, res := range results {
		allErrors = append(allErrors, res.Errors...)
		allWarnings = append(allWarnings, res.Warnings...)
		if len(res.Errors) > 0 {
			hasBlocking = true
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results":       results,
		"has_blocking":  hasBlocking,
		"error_count":   len(allErrors),
		"warning_count": len(allWarnings),
		"total_rules":   len(engineRules),
	})
}

func convertConditions(in []storage.RuleCondition) []rules.Condition {
	out := make([]rules.Condition, len(in))
	for i, c := range in {
		out[i] = rules.Condition{Field: c.Field, Operator: c.Operator, Value: c.Value}
	}
	return out
}

func convertActions(in []storage.RuleAction) []rules.Action {
	out := make([]rules.Action, len(in))
	for i, a := range in {
		out[i] = rules.Action{Type: a.Type, Params: a.Params}
	}
	return out
}

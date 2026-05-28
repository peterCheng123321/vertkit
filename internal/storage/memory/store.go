package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/vertkit/vertkit/internal/domain"
	"github.com/vertkit/vertkit/internal/storage"
)

// data holds all in-memory state. All store types share a pointer to this.
type data struct {
	mu            sync.RWMutex
	tenants       map[domain.TenantID]domain.Tenant
	accounts      map[string]domain.Account
	contacts      map[string]domain.Contact
	opportunities map[string]domain.Opportunity
	products      map[string]domain.Product
	rules         map[string]storage.Rule // key = tenantID + ":" + ruleID for simplicity
}

func newData() *data {
	return &data{
		tenants:       make(map[domain.TenantID]domain.Tenant),
		accounts:      make(map[string]domain.Account),
		contacts:      make(map[string]domain.Contact),
		opportunities: make(map[string]domain.Opportunity),
		products:      make(map[string]domain.Product),
		rules:         make(map[string]storage.Rule),
	}
}

func scopedKey(tenantID domain.TenantID, id string) string {
	return string(tenantID) + "\x00" + id
}

func accountKey(tenantID domain.TenantID, id domain.AccountID) string {
	return scopedKey(tenantID, string(id))
}

func contactKey(tenantID domain.TenantID, id domain.ContactID) string {
	return scopedKey(tenantID, string(id))
}

func opportunityKey(tenantID domain.TenantID, id domain.OpportunityID) string {
	return scopedKey(tenantID, string(id))
}

func productKey(tenantID domain.TenantID, id domain.ProductID) string {
	return scopedKey(tenantID, string(id))
}

// --- TenantStore implementation ---

type tenantStore struct{ d *data }

func (s *tenantStore) Create(ctx context.Context, t domain.Tenant) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	if _, exists := s.d.tenants[t.ID]; exists {
		return errors.New("tenant already exists")
	}
	s.d.tenants[t.ID] = t
	return nil
}

func (s *tenantStore) Get(ctx context.Context, id domain.TenantID) (*domain.Tenant, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	t, ok := s.d.tenants[id]
	if !ok {
		return nil, errors.New("tenant not found")
	}
	cp := t
	return &cp, nil
}

func (s *tenantStore) List(ctx context.Context) ([]*domain.Tenant, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Tenant, 0, len(s.d.tenants))
	for _, t := range s.d.tenants {
		cp := t
		out = append(out, &cp)
	}
	return out, nil
}

func (s *tenantStore) Update(ctx context.Context, t domain.Tenant) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	if _, exists := s.d.tenants[t.ID]; !exists {
		return errors.New("tenant not found")
	}
	s.d.tenants[t.ID] = t
	return nil
}

func (s *tenantStore) Delete(ctx context.Context, id domain.TenantID) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	delete(s.d.tenants, id)
	return nil
}

// --- AccountStore implementation ---

type accountStore struct{ d *data }

func (s *accountStore) Create(ctx context.Context, tenantID domain.TenantID, a domain.Account) error {
	if a.TenantID != tenantID {
		return errors.New("tenant mismatch on account")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := accountKey(tenantID, a.ID)
	if _, exists := s.d.accounts[key]; exists {
		return errors.New("account already exists")
	}
	s.d.accounts[key] = cloneAccount(a)
	return nil
}

func (s *accountStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) (*domain.Account, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	a, ok := s.d.accounts[accountKey(tenantID, id)]
	if !ok || a.TenantID != tenantID {
		return nil, errors.New("account not found")
	}
	cp := cloneAccount(a)
	return &cp, nil
}

func (s *accountStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Account, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Account, 0)
	for _, a := range s.d.accounts {
		if a.TenantID == tenantID {
			cp := cloneAccount(a)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *accountStore) Update(ctx context.Context, tenantID domain.TenantID, a domain.Account) error {
	if a.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := accountKey(tenantID, a.ID)
	existing, ok := s.d.accounts[key]
	if !ok || existing.TenantID != tenantID {
		return errors.New("account not found")
	}
	s.d.accounts[key] = cloneAccount(a)
	return nil
}

func (s *accountStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := accountKey(tenantID, id)
	a, ok := s.d.accounts[key]
	if !ok || a.TenantID != tenantID {
		return errors.New("account not found")
	}
	delete(s.d.accounts, key)
	return nil
}

// --- ContactStore ---

type contactStore struct{ d *data }

func (s *contactStore) Create(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error {
	if c.TenantID != tenantID {
		return errors.New("tenant mismatch on contact")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := contactKey(tenantID, c.ID)
	if _, exists := s.d.contacts[key]; exists {
		return errors.New("contact already exists")
	}
	s.d.contacts[key] = cloneContact(c)
	return nil
}

func (s *contactStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) (*domain.Contact, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	c, ok := s.d.contacts[contactKey(tenantID, id)]
	if !ok || c.TenantID != tenantID {
		return nil, errors.New("contact not found")
	}
	cp := cloneContact(c)
	return &cp, nil
}

func (s *contactStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Contact, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Contact, 0)
	for _, c := range s.d.contacts {
		if c.TenantID == tenantID {
			cp := cloneContact(c)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *contactStore) ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Contact, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Contact, 0)
	for _, c := range s.d.contacts {
		if c.TenantID == tenantID && c.AccountID == accountID {
			cp := cloneContact(c)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *contactStore) Update(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error {
	if c.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := contactKey(tenantID, c.ID)
	existing, ok := s.d.contacts[key]
	if !ok || existing.TenantID != tenantID {
		return errors.New("contact not found")
	}
	s.d.contacts[key] = cloneContact(c)
	return nil
}

func (s *contactStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := contactKey(tenantID, id)
	c, ok := s.d.contacts[key]
	if !ok || c.TenantID != tenantID {
		return errors.New("contact not found")
	}
	delete(s.d.contacts, key)
	return nil
}

// --- OpportunityStore ---

type opportunityStore struct{ d *data }

func (s *opportunityStore) Create(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error {
	if o.TenantID != tenantID {
		return errors.New("tenant mismatch on opportunity")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := opportunityKey(tenantID, o.ID)
	if _, exists := s.d.opportunities[key]; exists {
		return errors.New("opportunity already exists")
	}
	s.d.opportunities[key] = cloneOpportunity(o)
	return nil
}

func (s *opportunityStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) (*domain.Opportunity, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	o, ok := s.d.opportunities[opportunityKey(tenantID, id)]
	if !ok || o.TenantID != tenantID {
		return nil, errors.New("opportunity not found")
	}
	cp := cloneOpportunity(o)
	return &cp, nil
}

func (s *opportunityStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Opportunity, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Opportunity, 0)
	for _, o := range s.d.opportunities {
		if o.TenantID == tenantID {
			cp := cloneOpportunity(o)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *opportunityStore) ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Opportunity, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Opportunity, 0)
	for _, o := range s.d.opportunities {
		if o.TenantID == tenantID && o.AccountID == accountID {
			cp := cloneOpportunity(o)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *opportunityStore) Update(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error {
	if o.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := opportunityKey(tenantID, o.ID)
	existing, ok := s.d.opportunities[key]
	if !ok || existing.TenantID != tenantID {
		return errors.New("opportunity not found")
	}
	s.d.opportunities[key] = cloneOpportunity(o)
	return nil
}

func (s *opportunityStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := opportunityKey(tenantID, id)
	o, ok := s.d.opportunities[key]
	if !ok || o.TenantID != tenantID {
		return errors.New("opportunity not found")
	}
	delete(s.d.opportunities, key)
	return nil
}

// --- ProductStore ---

type productStore struct{ d *data }

func (s *productStore) Create(ctx context.Context, tenantID domain.TenantID, p domain.Product) error {
	if p.TenantID != tenantID {
		return errors.New("tenant mismatch on product")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := productKey(tenantID, p.ID)
	if _, exists := s.d.products[key]; exists {
		return errors.New("product already exists")
	}
	s.d.products[key] = cloneProduct(p)
	return nil
}

func (s *productStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) (*domain.Product, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	p, ok := s.d.products[productKey(tenantID, id)]
	if !ok || p.TenantID != tenantID {
		return nil, errors.New("product not found")
	}
	cp := cloneProduct(p)
	return &cp, nil
}

func (s *productStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Product, 0)
	for _, p := range s.d.products {
		if p.TenantID == tenantID {
			cp := cloneProduct(p)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *productStore) ListActive(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*domain.Product, 0)
	for _, p := range s.d.products {
		if p.TenantID == tenantID && p.Active {
			cp := cloneProduct(p)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *productStore) Update(ctx context.Context, tenantID domain.TenantID, p domain.Product) error {
	if p.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := productKey(tenantID, p.ID)
	existing, ok := s.d.products[key]
	if !ok || existing.TenantID != tenantID {
		return errors.New("product not found")
	}
	s.d.products[key] = cloneProduct(p)
	return nil
}

func (s *productStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := productKey(tenantID, id)
	p, ok := s.d.products[key]
	if !ok || p.TenantID != tenantID {
		return errors.New("product not found")
	}
	delete(s.d.products, key)
	return nil
}

// --- RuleStore (business rules) ---

type ruleStore struct{ d *data }

func ruleKey(tenantID domain.TenantID, id string) string {
	return string(tenantID) + ":" + id
}

func (s *ruleStore) Create(ctx context.Context, tenantID domain.TenantID, r storage.Rule) error {
	if r.TenantID != tenantID {
		return errors.New("tenant mismatch on rule")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := ruleKey(tenantID, r.ID)
	if _, exists := s.d.rules[key]; exists {
		return errors.New("rule already exists")
	}
	s.d.rules[key] = cloneRule(r)
	return nil
}

func (s *ruleStore) Get(ctx context.Context, tenantID domain.TenantID, id string) (*storage.Rule, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	key := ruleKey(tenantID, id)
	r, ok := s.d.rules[key]
	if !ok || r.TenantID != tenantID {
		return nil, errors.New("rule not found")
	}
	cp := cloneRule(r)
	return &cp, nil
}

func (s *ruleStore) List(ctx context.Context, tenantID domain.TenantID) ([]*storage.Rule, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*storage.Rule, 0)
	for _, r := range s.d.rules {
		if r.TenantID == tenantID {
			cp := cloneRule(r)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *ruleStore) ListActive(ctx context.Context, tenantID domain.TenantID, entityType string) ([]*storage.Rule, error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
	out := make([]*storage.Rule, 0)
	for _, r := range s.d.rules {
		if r.TenantID == tenantID && r.IsActive && (entityType == "" || r.EntityType == entityType) {
			cp := cloneRule(r)
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (s *ruleStore) Update(ctx context.Context, tenantID domain.TenantID, r storage.Rule) error {
	if r.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := ruleKey(tenantID, r.ID)
	existing, ok := s.d.rules[key]
	if !ok || existing.TenantID != tenantID {
		return errors.New("rule not found")
	}
	s.d.rules[key] = cloneRule(r)
	return nil
}

func (s *ruleStore) Delete(ctx context.Context, tenantID domain.TenantID, id string) error {
	s.d.mu.Lock()
	defer s.d.mu.Unlock()
	key := ruleKey(tenantID, id)
	r, ok := s.d.rules[key]
	if !ok || r.TenantID != tenantID {
		return errors.New("rule not found")
	}
	delete(s.d.rules, key)
	return nil
}

// NewStores returns a Stores struct with all in-memory implementations.
func NewStores() *storage.Stores {
	d := newData()
	return &storage.Stores{
		Tenants:       &tenantStore{d: d},
		Accounts:      &accountStore{d: d},
		Contacts:      &contactStore{d: d},
		Opportunities: &opportunityStore{d: d},
		Products:      &productStore{d: d},
		Rules:         &ruleStore{d: d},
	}
}

func cloneAccount(a domain.Account) domain.Account {
	a.CustomFields = cloneMap(a.CustomFields)
	if a.AnnualRevenue != nil {
		annualRevenue := *a.AnnualRevenue
		a.AnnualRevenue = &annualRevenue
	}
	return a
}

func cloneContact(c domain.Contact) domain.Contact {
	c.CustomFields = cloneMap(c.CustomFields)
	return c
}

func cloneOpportunity(o domain.Opportunity) domain.Opportunity {
	o.CustomFields = cloneMap(o.CustomFields)
	if o.CloseDate != nil {
		closeDate := *o.CloseDate
		o.CloseDate = &closeDate
	}
	if o.ExpectedRevenue != nil {
		expectedRevenue := *o.ExpectedRevenue
		o.ExpectedRevenue = &expectedRevenue
	}
	return o
}

func cloneProduct(p domain.Product) domain.Product {
	p.CustomFields = cloneMap(p.CustomFields)
	return p
}

func cloneRule(r storage.Rule) storage.Rule {
	r.Conditions = cloneConditions(r.Conditions)
	r.Actions = cloneActions(r.Actions)
	return r
}

func cloneConditions(in []storage.RuleCondition) []storage.RuleCondition {
	if in == nil {
		return nil
	}
	out := make([]storage.RuleCondition, len(in))
	for i, condition := range in {
		condition.Value = cloneAny(condition.Value)
		out[i] = condition
	}
	return out
}

func cloneActions(in []storage.RuleAction) []storage.RuleAction {
	if in == nil {
		return nil
	}
	out := make([]storage.RuleAction, len(in))
	for i, action := range in {
		action.Params = cloneMap(action.Params)
		out[i] = action
	}
	return out
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneAny(value)
	}
	return out
}

func cloneAny(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneMap(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneAny(item)
		}
		return out
	default:
		return value
	}
}

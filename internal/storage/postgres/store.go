package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vertkit/vertkit/internal/domain"
	"github.com/vertkit/vertkit/internal/storage"
)

type tenantStore struct{ db *sql.DB }
type accountStore struct{ db *sql.DB }
type contactStore struct{ db *sql.DB }
type opportunityStore struct{ db *sql.DB }
type productStore struct{ db *sql.DB }
type ruleStore struct{ db *sql.DB }

//go:embed schema.sql
var schemaSQL string

// SchemaSQL returns the Foundation v1 Postgres schema.
func SchemaSQL() string {
	return schemaSQL
}

// ApplySchema applies the idempotent Foundation v1 schema.
func ApplySchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schemaSQL)
	return err
}

// NewStores wires Postgres implementations behind the storage interfaces.
func NewStores(db *sql.DB) *storage.Stores {
	return &storage.Stores{
		Tenants:       &tenantStore{db: db},
		Accounts:      &accountStore{db: db},
		Contacts:      &contactStore{db: db},
		Opportunities: &opportunityStore{db: db},
		Products:      &productStore{db: db},
		Rules:         &ruleStore{db: db},
	}
}

func (s *tenantStore) Create(ctx context.Context, t domain.Tenant) error {
	return insertTenant(ctx, s.db, t)
}

func (s *tenantStore) Get(ctx context.Context, id domain.TenantID) (*domain.Tenant, error) {
	return getJSON[domain.Tenant](ctx, s.db, "SELECT data FROM tenants WHERE id = $1", id)
}

func (s *tenantStore) List(ctx context.Context) ([]*domain.Tenant, error) {
	return listJSON[domain.Tenant](ctx, s.db, "SELECT data FROM tenants ORDER BY id")
}

func (s *tenantStore) Update(ctx context.Context, t domain.Tenant) error {
	return updateTenant(ctx, s.db, t)
}

func (s *tenantStore) Delete(ctx context.Context, id domain.TenantID) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", id)
	return requireChanged(res, err, "tenant not found")
}

func (s *accountStore) Create(ctx context.Context, tenantID domain.TenantID, a domain.Account) error {
	if a.TenantID != tenantID {
		return errors.New("tenant mismatch on account")
	}
	return insertEntity(ctx, s.db, "accounts", tenantID, string(a.ID), a, a.CreatedAt, a.UpdatedAt)
}

func (s *accountStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) (*domain.Account, error) {
	return getEntity[domain.Account](ctx, s.db, "accounts", tenantID, string(id))
}

func (s *accountStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Account, error) {
	return listEntities[domain.Account](ctx, s.db, "accounts", tenantID)
}

func (s *accountStore) Update(ctx context.Context, tenantID domain.TenantID, a domain.Account) error {
	if a.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	return updateEntity(ctx, s.db, "accounts", tenantID, string(a.ID), a, a.UpdatedAt)
}

func (s *accountStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.AccountID) error {
	return deleteEntity(ctx, s.db, "accounts", tenantID, string(id), "account not found")
}

func (s *contactStore) Create(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error {
	if c.TenantID != tenantID {
		return errors.New("tenant mismatch on contact")
	}
	return insertContact(ctx, s.db, tenantID, c)
}

func (s *contactStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) (*domain.Contact, error) {
	return getEntity[domain.Contact](ctx, s.db, "contacts", tenantID, string(id))
}

func (s *contactStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Contact, error) {
	return listEntities[domain.Contact](ctx, s.db, "contacts", tenantID)
}

func (s *contactStore) ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Contact, error) {
	return listJSON[domain.Contact](ctx, s.db, "SELECT data FROM contacts WHERE tenant_id = $1 AND account_id = $2 ORDER BY id", tenantID, accountID)
}

func (s *contactStore) Update(ctx context.Context, tenantID domain.TenantID, c domain.Contact) error {
	if c.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	return updateContact(ctx, s.db, tenantID, c)
}

func (s *contactStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.ContactID) error {
	return deleteEntity(ctx, s.db, "contacts", tenantID, string(id), "contact not found")
}

func (s *opportunityStore) Create(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error {
	if o.TenantID != tenantID {
		return errors.New("tenant mismatch on opportunity")
	}
	return insertOpportunity(ctx, s.db, tenantID, o)
}

func (s *opportunityStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) (*domain.Opportunity, error) {
	return getEntity[domain.Opportunity](ctx, s.db, "opportunities", tenantID, string(id))
}

func (s *opportunityStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Opportunity, error) {
	return listEntities[domain.Opportunity](ctx, s.db, "opportunities", tenantID)
}

func (s *opportunityStore) ListByAccount(ctx context.Context, tenantID domain.TenantID, accountID domain.AccountID) ([]*domain.Opportunity, error) {
	return listJSON[domain.Opportunity](ctx, s.db, "SELECT data FROM opportunities WHERE tenant_id = $1 AND account_id = $2 ORDER BY id", tenantID, accountID)
}

func (s *opportunityStore) Update(ctx context.Context, tenantID domain.TenantID, o domain.Opportunity) error {
	if o.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	return updateOpportunity(ctx, s.db, tenantID, o)
}

func (s *opportunityStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.OpportunityID) error {
	return deleteEntity(ctx, s.db, "opportunities", tenantID, string(id), "opportunity not found")
}

func (s *productStore) Create(ctx context.Context, tenantID domain.TenantID, p domain.Product) error {
	if p.TenantID != tenantID {
		return errors.New("tenant mismatch on product")
	}
	return insertEntity(ctx, s.db, "products", tenantID, string(p.ID), p, p.CreatedAt, p.UpdatedAt)
}

func (s *productStore) Get(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) (*domain.Product, error) {
	return getEntity[domain.Product](ctx, s.db, "products", tenantID, string(id))
}

func (s *productStore) List(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error) {
	return listEntities[domain.Product](ctx, s.db, "products", tenantID)
}

func (s *productStore) ListActive(ctx context.Context, tenantID domain.TenantID) ([]*domain.Product, error) {
	return listJSON[domain.Product](ctx, s.db, "SELECT data FROM products WHERE tenant_id = $1 AND (data->>'active')::boolean = true ORDER BY id", tenantID)
}

func (s *productStore) Update(ctx context.Context, tenantID domain.TenantID, p domain.Product) error {
	if p.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	return updateEntity(ctx, s.db, "products", tenantID, string(p.ID), p, p.UpdatedAt)
}

func (s *productStore) Delete(ctx context.Context, tenantID domain.TenantID, id domain.ProductID) error {
	return deleteEntity(ctx, s.db, "products", tenantID, string(id), "product not found")
}

func (s *ruleStore) Create(ctx context.Context, tenantID domain.TenantID, r storage.Rule) error {
	if r.TenantID != tenantID {
		return errors.New("tenant mismatch on rule")
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, "INSERT INTO rules (tenant_id, id, entity_type, is_active, data, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)", tenantID, r.ID, r.EntityType, r.IsActive, data, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *ruleStore) Get(ctx context.Context, tenantID domain.TenantID, id string) (*storage.Rule, error) {
	return getEntity[storage.Rule](ctx, s.db, "rules", tenantID, id)
}

func (s *ruleStore) List(ctx context.Context, tenantID domain.TenantID) ([]*storage.Rule, error) {
	return listEntities[storage.Rule](ctx, s.db, "rules", tenantID)
}

func (s *ruleStore) ListActive(ctx context.Context, tenantID domain.TenantID, entityType string) ([]*storage.Rule, error) {
	return listJSON[storage.Rule](ctx, s.db, "SELECT data FROM rules WHERE tenant_id = $1 AND is_active = true AND ($2 = '' OR entity_type = $2) ORDER BY id", tenantID, entityType)
}

func (s *ruleStore) Update(ctx context.Context, tenantID domain.TenantID, r storage.Rule) error {
	if r.TenantID != tenantID {
		return errors.New("tenant mismatch")
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, "UPDATE rules SET entity_type = $3, is_active = $4, data = $5, updated_at = $6 WHERE tenant_id = $1 AND id = $2", tenantID, r.ID, r.EntityType, r.IsActive, data, r.UpdatedAt)
	return requireChanged(res, err, "rule not found")
}

func (s *ruleStore) Delete(ctx context.Context, tenantID domain.TenantID, id string) error {
	return deleteEntity(ctx, s.db, "rules", tenantID, id, "rule not found")
}

func insertTenant(ctx context.Context, db *sql.DB, t domain.Tenant) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO tenants (id, data, created_at, updated_at) VALUES ($1, $2, $3, $4)", t.ID, data, t.CreatedAt, t.UpdatedAt)
	return err
}

func updateTenant(ctx context.Context, db *sql.DB, t domain.Tenant) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, "UPDATE tenants SET data = $2, updated_at = $3 WHERE id = $1", t.ID, data, t.UpdatedAt)
	return requireChanged(res, err, "tenant not found")
}

func insertContact(ctx context.Context, db *sql.DB, tenantID domain.TenantID, c domain.Contact) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO contacts (tenant_id, id, account_id, data, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", tenantID, c.ID, nullableString(c.AccountID), data, c.CreatedAt, c.UpdatedAt)
	return err
}

func updateContact(ctx context.Context, db *sql.DB, tenantID domain.TenantID, c domain.Contact) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, "UPDATE contacts SET account_id = $3, data = $4, updated_at = $5 WHERE tenant_id = $1 AND id = $2", tenantID, c.ID, nullableString(c.AccountID), data, c.UpdatedAt)
	return requireChanged(res, err, "contact not found")
}

func insertOpportunity(ctx context.Context, db *sql.DB, tenantID domain.TenantID, o domain.Opportunity) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO opportunities (tenant_id, id, account_id, data, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)", tenantID, o.ID, o.AccountID, data, o.CreatedAt, o.UpdatedAt)
	return err
}

func updateOpportunity(ctx context.Context, db *sql.DB, tenantID domain.TenantID, o domain.Opportunity) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, "UPDATE opportunities SET account_id = $3, data = $4, updated_at = $5 WHERE tenant_id = $1 AND id = $2", tenantID, o.ID, o.AccountID, data, o.UpdatedAt)
	return requireChanged(res, err, "opportunity not found")
}

func insertEntity(ctx context.Context, db *sql.DB, table string, tenantID domain.TenantID, id string, value any, createdAt any, updatedAt any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (tenant_id, id, data, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", table), tenantID, id, data, createdAt, updatedAt)
	return err
}

func updateEntity(ctx context.Context, db *sql.DB, table string, tenantID domain.TenantID, id string, value any, updatedAt any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET data = $3, updated_at = $4 WHERE tenant_id = $1 AND id = $2", table), tenantID, id, data, updatedAt)
	return requireChanged(res, err, fmt.Sprintf("%s not found", table))
}

func getEntity[T any](ctx context.Context, db *sql.DB, table string, tenantID domain.TenantID, id string) (*T, error) {
	return getJSON[T](ctx, db, fmt.Sprintf("SELECT data FROM %s WHERE tenant_id = $1 AND id = $2", table), tenantID, id)
}

func listEntities[T any](ctx context.Context, db *sql.DB, table string, tenantID domain.TenantID) ([]*T, error) {
	return listJSON[T](ctx, db, fmt.Sprintf("SELECT data FROM %s WHERE tenant_id = $1 ORDER BY id", table), tenantID)
}

func deleteEntity(ctx context.Context, db *sql.DB, table string, tenantID domain.TenantID, id string, notFound string) error {
	res, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND id = $2", table), tenantID, id)
	return requireChanged(res, err, notFound)
}

func getJSON[T any](ctx context.Context, db *sql.DB, query string, args ...any) (*T, error) {
	var data []byte
	if err := db.QueryRowContext(ctx, query, args...).Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func listJSON[T any](ctx context.Context, db *sql.DB, query string, args ...any) ([]*T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*T{}
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		out = append(out, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func requireChanged(res sql.Result, err error, notFound string) error {
	if err != nil {
		return err
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if changed == 0 {
		return errors.New(notFound)
	}
	return nil
}

func nullableString[T ~string](value T) any {
	if value == "" {
		return nil
	}
	return string(value)
}

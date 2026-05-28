CREATE TABLE IF NOT EXISTS tenants (
    id text PRIMARY KEY,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    tenant_id text NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    id text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS contacts (
    tenant_id text NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    id text NOT NULL,
    account_id text,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS opportunities (
    tenant_id text NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    id text NOT NULL,
    account_id text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS products (
    tenant_id text NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    id text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS rules (
    tenant_id text NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    id text NOT NULL,
    entity_type text NOT NULL,
    is_active boolean NOT NULL DEFAULT false,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS contacts_tenant_account_idx ON contacts (tenant_id, account_id);
CREATE INDEX IF NOT EXISTS opportunities_tenant_account_idx ON opportunities (tenant_id, account_id);
CREATE INDEX IF NOT EXISTS products_tenant_active_idx ON products (tenant_id, ((data->>'active')::boolean));
CREATE INDEX IF NOT EXISTS rules_tenant_active_entity_idx ON rules (tenant_id, is_active, entity_type);

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE opportunities ENABLE ROW LEVEL SECURITY;
ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE rules ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_tenants ON tenants;
CREATE POLICY tenant_isolation_tenants ON tenants
    USING (id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (id = current_setting('vertkit.tenant_id', true));

DROP POLICY IF EXISTS tenant_isolation_accounts ON accounts;
CREATE POLICY tenant_isolation_accounts ON accounts
    USING (tenant_id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (tenant_id = current_setting('vertkit.tenant_id', true));

DROP POLICY IF EXISTS tenant_isolation_contacts ON contacts;
CREATE POLICY tenant_isolation_contacts ON contacts
    USING (tenant_id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (tenant_id = current_setting('vertkit.tenant_id', true));

DROP POLICY IF EXISTS tenant_isolation_opportunities ON opportunities;
CREATE POLICY tenant_isolation_opportunities ON opportunities
    USING (tenant_id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (tenant_id = current_setting('vertkit.tenant_id', true));

DROP POLICY IF EXISTS tenant_isolation_products ON products;
CREATE POLICY tenant_isolation_products ON products
    USING (tenant_id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (tenant_id = current_setting('vertkit.tenant_id', true));

DROP POLICY IF EXISTS tenant_isolation_rules ON rules;
CREATE POLICY tenant_isolation_rules ON rules
    USING (tenant_id = current_setting('vertkit.tenant_id', true))
    WITH CHECK (tenant_id = current_setting('vertkit.tenant_id', true));

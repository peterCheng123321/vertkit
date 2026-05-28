# VertKit

**The open-source foundation for the next generation of global CRM and ERP systems.**

VertKit is designed so that developers and ISVs building vertical CRM/ERP products can start with a clean, modern, globally-aware core — then let their agents (Claude, Cursor, custom agents, etc.) safely extend and customize it.

## Philosophy

- **Global from day one**: Every monetary value carries explicit currency. Addresses are structured for country-specific rules. Tenants have locale + currency defaults. No retrofitting internationalization later.
- **Agent-native**: The codebase is deliberately small, obvious, and well-typed so that LLMs can understand the entire domain model and safely generate extensions, custom fields, workflows, and vertical modules.
- **Multi-tenant by construction**: Tenant isolation is enforced at the storage interface level. Cross-tenant leaks are impossible by design.
- **Minimal core, powerful extensions**: We ship the essential entities (Account, Contact, Opportunity, Product) + the primitives (Money, Address, Tenant). Everything else — custom objects, complex processes, industry verticals — is meant to be added by the implementer or their agents.
- **Storage is pluggable**: Today we ship an in-memory store for instant start. Tomorrow: PostgreSQL, MySQL, or even connectors to existing systems.

## Current Status (First Slice)

This is the **core domain models + basic multi-tenant CRUD API** milestone.

**Entities included**:
- `Tenant`
- `Account`
- `Contact`
- `Opportunity`
- `Product`

**Global primitives**:
- `Money` (amount in minor units + ISO currency, no floats)
- `Address` (structured with CountryCode)

**What works today**:
- Full CRUD + listing for tenants and all core entities
- Tenant-scoped storage boundaries, including duplicate-safe entity IDs across tenants
- Optional service-token protection for tenant-scoped API calls
- In-memory storage for instant local feedback
- Postgres storage via `DATABASE_URL`, with an idempotent RLS-ready schema
- OpenAPI 3.1 contract in `openapi.yaml`
- TypeScript agent backend in `apps/agent` with a tool registry, permission gates, audit log, OpenAI-compatible provider adapter, CRM API tools, CLI, and HTTP job endpoint

**What is intentionally not here yet** (by design):
- Full end-user authentication / authorization
- Durable workflow persistence, queues, or scheduler
- Frontend (React/TS layer coming in a later slice)
- Custom field type system (we have a simple `map[string]any` escape hatch for now)
- Invoicing, orders, inventory, or full ERP flows

## Quick Start

```bash
cd vertkit
export VERTKIT_SERVICE_TOKEN=dev-token
go run ./cmd/server
```

Set `DATABASE_URL` to use Postgres instead of in-memory storage:

```bash
export DATABASE_URL='postgres://user:pass@localhost:5432/vertkit?sslmode=disable'
export VERTKIT_SERVICE_TOKEN=dev-token
go run ./cmd/server
```

In another terminal:

```bash
TOKEN=dev-token

# 1. Create a tenant
curl -X POST http://localhost:8080/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acme",
    "name": "Acme Global Inc",
    "default_currency": "USD",
    "default_locale": "en-US"
  }'

# 2. Create an account inside that tenant
curl -X POST http://localhost:8080/tenants/acme/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acc_001",
    "name": "Globex Corporation",
    "industry": "Software"
  }'

# 3. Create a contact
curl -X POST http://localhost:8080/tenants/acme/contacts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "ct_001",
    "account_id": "acc_001",
    "first_name": "Alice",
    "last_name": "Chen",
    "email": "alice@globex.com"
  }'

# 4. Create an opportunity with proper Money
curl -X POST http://localhost:8080/tenants/acme/opportunities \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "opp_001",
    "account_id": "acc_001",
    "name": "Enterprise License 2026",
    "amount": 25000000,
    "currency": "USD"
  }'

# 5. List accounts for the tenant
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/tenants/acme/accounts
```

Health check: `curl http://localhost:8080/health`

## TypeScript Agent Backend

VertKit now includes an original TypeScript agent runtime. It is not LangGraph. It is a small VertKit-owned harness for CRM automation and future code-building workflows:

- `ToolRegistry` exposes typed CRM tools.
- `PermissionPolicy` allows, denies, or pauses risky actions for approval.
- `AgentRuntime` runs the provider/tool loop and writes an audit trail.
- `OpenAICompatibleProvider` talks to any chat-completions-compatible model that returns JSON decisions.
- `crmApiTools` call the Go CRM API with tenant context and service-token auth.
- `createAgentServer` exposes `/health`, `/agent/tools`, and `/agent/jobs`.

Run the agent CLI:

```bash
npm run agent:start -- tools
npm run agent:start -- run "Summarize accounts for this tenant"
```

Run the HTTP agent service:

```bash
export VERTKIT_CRM_BASE_URL=http://127.0.0.1:8080
export VERTKIT_SERVICE_TOKEN=dev-token
export VERTKIT_AGENT_BASE_URL=https://api.openai.com/v1
export VERTKIT_AGENT_API_KEY=...
export VERTKIT_AGENT_MODEL=...

npm run agent:start -- serve --port 8787
```

Submit a job:

```bash
curl -X POST http://127.0.0.1:8787/agent/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme",
    "actor": "agent:http",
    "goal": "List accounts and identify missing contacts"
  }'
```

## How Agents Customize VertKit

The goal is that after `go run` (or later `npm install @vertkit/core` + a thin server), a developer can say to their agent:

> "Add a custom 'RenewalDate' field to Opportunity and a validation rule that prevents moving to closed-won if the renewal date is in the past."

Because the domain types are plain, the storage interfaces are narrow, and the HTTP layer is thin and predictable, agents have a high chance of producing correct, safe extensions instead of fighting a giant legacy codebase.

**Recommended patterns for agent-driven customization** (will be expanded in docs/):

1. Add fields via the existing `CustomFields map[string]any` on entities (fast).
2. Later: introduce typed extension registries.
3. Add new entity types by copying the pattern of Account/Contact (new domain struct + new store interface + new routes).
4. Never mutate core Money or Address logic — extend around them.

## Business Rules Engine (Current Slice)

VertKit now includes a simple but real **business rules engine** — the foundation for highly customizable logic without touching core code.

Rules are:
- Tenant-scoped and stored
- Defined with clear JSON conditions + actions
- Evaluated on demand (perfect for agents or UI before save)
- Designed so LLMs can generate correct rules easily

### Example Business Rules

**1. Large Deal Approval Gate (Money-aware + global)**
```json
{
  "id": "large_deal_approval",
  "name": "Large Deal Requires Approval",
  "entity_type": "opportunity",
  "conditions": [
    {"field": "amount", "operator": "gt", "value": 10000000},
    {"field": "currency", "operator": "eq", "value": "USD"}
  ],
  "actions": [
    {"type": "block", "params": {"message": "Deals over $100k USD require manager approval before advancing"}}
  ],
  "is_active": true
}
```

**2. Stage Gate Rule**
```json
{
  "id": "no_closed_won_low_probability",
  "name": "Cannot close won with low probability",
  "entity_type": "opportunity",
  "conditions": [
    {"field": "stage", "operator": "eq", "value": "closed-won"},
    {"field": "probability", "operator": "lt", "value": 70}
  ],
  "actions": [
    {"type": "block", "params": {"message": "Probability must be at least 70% to mark as closed-won"}}
  ],
  "is_active": true
}
```

**3. Global Currency / Region Rule**
```json
{
  "id": "emea_high_value_review",
  "name": "EMEA high-value deals need extra review",
  "entity_type": "opportunity",
  "conditions": [
    {"field": "amount", "operator": "gte", "value": 5000000},
    {"field": "custom_fields.region", "operator": "in", "value": ["EMEA", "EU", "UK"]}
  ],
  "actions": [
    {"type": "warn", "params": {"message": "High-value EMEA deal — finance review recommended"}}
  ],
  "is_active": true
}
```

### Using the Rules API

Create a rule:
```bash
curl -X POST http://localhost:8080/tenants/acme/rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @rule-large-deal.json
```

Evaluate rules against an opportunity snapshot (what agents/UI will call):
```bash
curl -X POST http://localhost:8080/tenants/acme/rules/evaluate/opportunity \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "opportunity": { ... full opportunity object ... },
    "operation": "update"
  }'
```

Response tells you `has_blocking`, warnings, and suggested changes.

This is the beginning of the "agent embedded" experience: agents can propose opportunities, then immediately run the tenant's active rules to see if the action is allowed.

## Project Layout

```
vertkit/
├── apps/agent/                 # TypeScript agent backend and CLI
├── cmd/server/main.go          # entrypoint
├── internal/
│   ├── domain/                 # the heart — pure, global-first types
│   │   ├── money.go
│   │   ├── address.go
│   │   ├── tenant.go
│   │   ├── account.go
│   │   ├── contact.go
│   │   ├── opportunity.go
│   │   └── product.go
│   ├── storage/
│   │   ├── storage.go          # the pluggable interfaces (key extension point)
│   │   ├── memory/store.go
│   │   └── postgres/           # Postgres store + RLS-ready schema
│   ├── api/
│   │   └── server.go           # minimal HTTP layer
│   └── rules/                  # business rules engine (conditions + actions)
│       ├── rule.go
│       └── engine.go
├── README.md
├── package.json
├── openapi.yaml
└── go.mod
```

## Roadmap (High Level)

1. **Foundation v1** (current) — tenant-safe CRUD, Postgres, OpenAPI
2. **Agent runtime v1** (current) — TypeScript tool registry, permission gates, audit log, CRM API tools, and HTTP job endpoint
3. Typed custom fields + validation hooks
4. Durable agent jobs with queueing, approvals, retries, and persisted audit events
5. Basic workflow / state machine primitives
6. TypeScript client + thin React reference UI
7. "create-vertkit" CLI / npm experience
8. Reference vertical modules (e.g. professional services, manufacturing, SaaS)

## Contributing & Philosophy

VertKit is meant to be the boring, correct, global foundation that hundreds of specialized CRM/ERP products can be built on top of — not another monolithic all-in-one.

If you are building a vertical solution and your agent is generating code against this codebase, please open issues with the prompts that worked well and the ones that didn't. Improving "agent ergonomics" is a first-class goal.

## License

MIT.

## Agent Skill / Plugin

VertKit includes a reusable Claude/Codex skill for future CRM work:

- Skill source: `skills/vertkit-crm-builder/`
- Claude plugin manifest: `.claude-plugin/plugin.json`
- Claude marketplace manifest: `.claude-plugin/marketplace.json`
- Codex plugin manifest: `.codex-plugin/plugin.json`

The skill teaches future agents the VertKit architecture, tenant-safety invariants,
core-vs-vertical module decisions, and the TypeScript agent runtime boundary.

---

**Status**: Early. The foundation is real and compiles cleanly. Everything beyond the current slice is still to be designed in the open.

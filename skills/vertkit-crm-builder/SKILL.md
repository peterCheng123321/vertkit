---
name: vertkit-crm-builder
description: Use when building or extending VertKit CRM/ERP systems, vertical CRM modules, tenant-scoped backend features, Postgres/OpenAPI persistence contracts, service-token integrations, business rules, or LangGraph agent workflows that operate on CRM data.
---

# VertKit CRM Builder

VertKit is a tenant-safe CRM/ERP backend foundation for vertical builders. Use the Go API and Postgres as the system of record; use LangGraph only for durable workflows that call the API.

## Start Here

1. Read repo truth first: `README.md`, `openapi.yaml`, and `internal/storage/storage.go`.
2. Classify the request:
   - **Core platform**: reusable CRM behavior. Read [references/backend-foundation.md](references/backend-foundation.md).
   - **Vertical CRM module**: industry-specific behavior. Read [references/vertical-modules.md](references/vertical-modules.md).
   - **Agent workflow / LangGraph**: orchestration, approvals, enrichment, or human review. Read [references/langgraph-boundary.md](references/langgraph-boundary.md).
3. Use TDD for behavior changes. Add tests before implementation.
4. Run `scripts/verify_vertkit.sh` before claiming completion.

## Non-Negotiable Invariants

- `tenant_id` is a security boundary, not a filter convenience.
- Tenant-scoped IDs are unique only within a tenant; storage keys and Postgres keys must include tenant ID.
- Validate tenant existence before tenant-scoped writes.
- Reject path/body tenant mismatches.
- Return JSON errors as `{"error":"message"}`.
- Keep memory storage for fast tests/dev; keep Postgres as the production path.
- Deep-copy mutable maps/slices at store boundaries.
- Update `openapi.yaml` when public API behavior changes.

## Default Architecture

- Go core API owns CRM records and authorization boundaries.
- Postgres owns durable business state and tenant-aware persistence.
- REST + OpenAPI is the public contract for apps, SDKs, agents, and LangGraph.
- Python LangGraph owns workflow state, checkpoints, retries, approvals, and traces. It must call VertKit APIs instead of writing CRM tables directly.

## Common Workflows

**Add a core CRM entity or capability**

Read [references/backend-foundation.md](references/backend-foundation.md). Implement domain type, storage interface, memory store, Postgres schema/store, API routes, OpenAPI contract, and tests together.

**Add vertical behavior**

Read [references/vertical-modules.md](references/vertical-modules.md). Prefer `custom_fields`, rules, and module-specific API layers before changing core structs.

**Add a rule/operator/action**

Keep rules agent-generatable. Field paths use JSON names like `amount`, `currency`, and `custom_fields.region`. Add regression tests for every new operator, path shape, or action.

**Add LangGraph automation**

Read [references/langgraph-boundary.md](references/langgraph-boundary.md). Treat LangGraph proposals as workflow outputs that call VertKit APIs with tenant context and service credentials.

## Completion Bar

Before finishing:

```bash
skills/vertkit-crm-builder/scripts/verify_vertkit.sh
```

If running from outside the repo, `cd` to the VertKit root first. If the skill is installed globally, use the installed script path and still run it from the project root.

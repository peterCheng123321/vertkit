# Backend Foundation Reference

Use this for core reusable CRM behavior: entities, stores, REST routes, OpenAPI, Postgres, service-token tenancy, and business rules.

## Implementation Shape

Core capability changes usually require all of these:

- Domain model in `internal/domain`.
- Storage interface in `internal/storage/storage.go`.
- Memory implementation in `internal/storage/memory`.
- Postgres schema/store in `internal/storage/postgres`.
- API handlers in `internal/api/server.go`.
- Public contract in `openapi.yaml`.
- Tests for domain, store, API, rules, schema, and OpenAPI as applicable.

Do not add only a handler or only a struct. A reusable backend feature is not real until the storage contract, API contract, and tests agree.

## Tenant Safety Checklist

- Every tenant-scoped type has `TenantID`.
- Every store method accepts `tenantID`.
- Every map key or SQL primary key includes tenant ID.
- Create/update rejects body tenant mismatch.
- Tenant-scoped API handlers verify the tenant exists before writes.
- Postgres tables use `(tenant_id, id)` primary keys for tenant-scoped entities.
- Postgres schema remains RLS-ready with `current_setting('vertkit.tenant_id', true)`.

## API Pattern

Use predictable REST:

- `POST /tenants/{tenant_id}/{entities}`
- `GET /tenants/{tenant_id}/{entities}`
- `GET /tenants/{tenant_id}/{entities}/{id}`
- `PUT /tenants/{tenant_id}/{entities}/{id}`
- `DELETE /tenants/{tenant_id}/{entities}/{id}`

Keep route behavior boring:

- `201` on create.
- `200` on get/update/list.
- `204` on delete.
- `400` for invalid input or reference mismatch.
- `401` for missing/invalid service token.
- `404` for missing tenant/entity.
- `409` for duplicate IDs.

## Store Pattern

Memory store:

- Use tenant-scoped keys.
- Copy structs on input and output.
- Deep-copy `map[string]any`, `[]any`, and pointer fields.
- Keep behavior aligned with Postgres store tests.

Postgres store:

- Store the whole entity JSON in `data jsonb` for v1 simplicity.
- Duplicate important query fields as columns only when listing/filtering needs them.
- Keep migrations idempotent.
- Do not bypass storage interfaces from API handlers.

## Money and Global Data

- Money uses minor units and explicit ISO currency.
- Do not expose float-based money paths in API/domain behavior.
- Currency comparisons in rules must work against JSON strings.
- Addresses should stay globally structured; add country-specific validation outside the core type until the validation design is explicit.

## Tests to Add

- Duplicate entity IDs can exist in different tenants.
- Cross-tenant reads/writes fail.
- Missing tenant writes fail.
- Store returns cannot mutate stored maps/slices.
- OpenAPI has every public route changed.
- Postgres schema has tenant keys and RLS-ready policies.

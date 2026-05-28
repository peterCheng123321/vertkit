# VertKit TypeScript Agent Backend

VertKit uses its own TypeScript agent layer instead of LangGraph. The goal is a small, auditable runtime that can automate CRM work without becoming the CRM system of record.

## Runtime Boundary

- The Go API owns tenant-scoped CRM data, validation, and storage.
- Postgres owns durable CRM state.
- The TypeScript agent backend owns task execution, tool routing, permission checks, provider calls, and audit events.
- Agents must call CRM APIs through tenant-scoped tools. They must not write directly to CRM tables.

## Components

- `AgentRuntime`: provider/tool loop with max-step protection.
- `ToolRegistry`: named tools with descriptions, risk levels, validation, and handlers.
- `PermissionPolicy`: allow, deny, or pause for approval before tool execution.
- `AuditLog`: append-only event log interface, with an in-memory implementation for the first slice.
- `OpenAICompatibleProvider`: adapter for chat-completions-compatible models that return JSON decisions.
- `crmApiTools`: CRM tools for health, tenants, accounts, contacts, opportunities, and rule evaluation.
- `createAgentServer`: HTTP service exposing health, tool inventory, and synchronous job execution.

## Decision Format

Providers return one JSON decision per step:

```json
{"type":"tool_call","toolName":"crm.accounts.list","input":{},"reason":"Need current accounts"}
```

or:

```json
{"type":"final","message":"Found 3 accounts. Two are missing contacts."}
```

## Permission Model

The default runtime uses risk-based policy:

- `read`: allowed
- `write`: approval required
- `external`: approval required
- `shell`: approval required

This makes the first runtime useful for read-heavy CRM automation while keeping write operations behind an approval boundary.

## HTTP API

`GET /health`

Returns agent service status.

`GET /agent/tools`

Returns the current tool inventory.

`POST /agent/jobs`

Runs a synchronous agent job:

```json
{
  "tenant_id": "acme",
  "actor": "agent:http",
  "goal": "List accounts and identify missing contacts",
  "max_steps": 8
}
```

The response includes `status`, `final_message`, `tool_results`, and any approval or denial state.

## Next Milestones

1. Persist audit events and jobs in Postgres.
2. Add resumable approvals for paused jobs.
3. Add queue-backed background execution.
4. Add idempotency keys for write tools.
5. Add stricter input schemas for all tools.
6. Add provider-specific adapters only behind the existing `AgentProvider` interface.

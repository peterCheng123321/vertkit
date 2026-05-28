# Vertical CRM Module Reference

Use this when adding industry-specific CRM behavior: real estate, healthcare, legal, recruiting, SaaS renewals, manufacturing, education, restaurants, or other vertical workflows.

## Classification Rule

Before changing core structs, decide:

- **Reusable across most CRMs**: core platform candidate.
- **Specific to one industry or workflow**: vertical module candidate.
- **Simple extra field**: use `custom_fields`.
- **Validation/approval**: use rules first.
- **Long-running process**: consider LangGraph workflow after core API support exists.

Default to module-specific behavior unless the concept is obviously foundational.

## Preferred Extension Order

1. `custom_fields` for simple metadata.
2. Business rules for validation, warnings, blocks, and suggested field changes.
3. Module-specific API routes/services if behavior becomes repeated or complex.
4. New core entity only when multiple verticals need the same primitive.

## Good Vertical Features

Examples that should usually stay vertical/module-specific:

- SaaS renewal date, ARR band, expansion health.
- Legal matter type, jurisdiction, opposing party.
- Real-estate listing status, property class, showing window.
- Recruiting candidate stage, interview loop, source channel.
- Restaurant supplier, menu item lifecycle, ingredient cost.

Examples that may become core:

- Activity/task primitives.
- Notes and attachments.
- Ownership and assignment.
- Audit events.
- Lifecycle/state-machine support.

## Rules for Agent-Built Modules

- Keep module boundaries obvious.
- Do not weaken tenant isolation to make module work easier.
- Avoid adding dozens of fields to core entities.
- If a module needs a new entity, give it its own store/API and OpenAPI section.
- Write examples in README/docs so future agents know the intended usage.

## Acceptance Criteria

A vertical module is good when:

- A builder can use it without reading unrelated internals.
- Core CRM entities remain small.
- Tenant boundaries are preserved.
- The module can be disabled or ignored by another vertical.
- Tests prove its primary workflow and failure cases.

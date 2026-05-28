# LangGraph Boundary Reference

Use this when adding agent workflows, approvals, enrichment, intake, automation, or human-in-the-loop flows.

## Architecture Rule

LangGraph is the workflow runtime, not the CRM backend.

- VertKit Go API owns CRM business records.
- Postgres owns durable CRM state.
- LangGraph owns workflow checkpoints, retries, thread state, tool calls, and human review.
- LangGraph commits CRM changes by calling VertKit REST APIs with tenant context.

Never let graph state become the only source for an Account, Contact, Opportunity, Product, Rule, or Tenant.

## Good LangGraph Use Cases

- Lead intake that asks follow-up questions and creates Account/Contact/Opportunity records.
- Approval workflow for large deals, discounts, or stage changes.
- Data enrichment that proposes updates and asks for review.
- Deduplication workflow that suggests merges.
- Background workflow that evaluates rules before committing a change.

## Required Integration Contract

Each workflow call must carry:

- Tenant ID.
- Service credential.
- Actor/user context when available.
- Idempotency key for commit operations when retries are possible.
- Trace/request ID so Go API logs and LangGraph runs can be correlated.

LangGraph should:

- Fetch current state through VertKit APIs.
- Produce proposed changes.
- Run rules/evaluation APIs before commit when relevant.
- Commit through VertKit APIs.
- Store workflow decisions/checkpoints in its own persistence.

## Anti-Patterns

- Writing directly to VertKit Postgres tables from LangGraph.
- Using checkpoint state as the CRM record store.
- Letting LLM-generated JSON bypass domain/API validation.
- Running workflows without tenant context.
- Treating failed API writes as successful graph completion.

## First LangGraph Milestone

After Foundation v1, build a separate Python service:

- FastAPI or LangGraph Platform endpoint for workflow runs.
- Postgres checkpointer/store for workflow state.
- VertKit API client generated from or aligned with `openapi.yaml`.
- Service-token configuration.
- Tests with a mocked VertKit API first, then integration tests.

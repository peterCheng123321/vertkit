import test from "node:test";
import assert from "node:assert/strict";

import { InMemoryAuditLog } from "./audit/AuditLog.ts";
import { createAgentServer } from "./server.ts";
import { AllowAllPolicy } from "./runtime/PermissionPolicy.ts";
import { AgentRuntime } from "./runtime/AgentRuntime.ts";
import { ToolRegistry } from "./runtime/ToolRegistry.ts";
import { ScriptedProvider } from "./providers/ScriptedProvider.ts";

test("server exposes health and tool inventory endpoints", async () => {
  const registry = new ToolRegistry();
  registry.register({
    name: "crm.health",
    description: "Checks CRM backend health",
    risk: "read",
    handler: async () => ({ content: { ok: true } }),
  });
  const runtime = new AgentRuntime({
    provider: new ScriptedProvider([{ type: "final", message: "idle" }]),
    registry,
    policy: new AllowAllPolicy(),
    auditLog: new InMemoryAuditLog(),
  });
  const server = createAgentServer({ runtime, registry });

  await server.listen(0);
  try {
    const baseUrl = `http://127.0.0.1:${server.port()}`;
    const health = await fetch(`${baseUrl}/health`);
    assert.equal(health.status, 200);
    assert.deepEqual(await health.json(), {
      ok: true,
      service: "vertkit-agent",
    });

    const tools = await fetch(`${baseUrl}/agent/tools`);
    assert.equal(tools.status, 200);
    assert.deepEqual(await tools.json(), {
      tools: [
        {
          name: "crm.health",
          description: "Checks CRM backend health",
          risk: "read",
        },
      ],
    });
  } finally {
    await server.close();
  }
});

test("server runs an agent job with tenant scope", async () => {
  const registry = new ToolRegistry();
  registry.register({
    name: "crm.accounts.list",
    description: "Lists accounts",
    risk: "read",
    handler: async (ctx) => ({
      content: [{ id: "acc_1", tenant_id: ctx.tenantId, name: "Acme" }],
    }),
  });
  const runtime = new AgentRuntime({
    provider: new ScriptedProvider([
      { type: "tool_call", toolName: "crm.accounts.list", input: {} },
      { type: "final", message: "Found 1 account." },
    ]),
    registry,
    policy: new AllowAllPolicy(),
    auditLog: new InMemoryAuditLog(),
  });
  const server = createAgentServer({ runtime, registry });

  await server.listen(0);
  try {
    const response = await fetch(`http://127.0.0.1:${server.port()}/agent/jobs`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        tenant_id: "tenant_1",
        actor: "agent:test",
        goal: "Summarize accounts",
      }),
    });

    assert.equal(response.status, 200);
    const body = await response.json();
    assert.equal(body.status, "completed");
    assert.equal(body.final_message, "Found 1 account.");
    assert.deepEqual(body.tool_results[0].content, [
      { id: "acc_1", tenant_id: "tenant_1", name: "Acme" },
    ]);
  } finally {
    await server.close();
  }
});

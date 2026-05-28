import test from "node:test";
import assert from "node:assert/strict";

import { AgentRuntime } from "./AgentRuntime.ts";
import { InMemoryAuditLog } from "../audit/AuditLog.ts";
import { AllowAllPolicy, StaticPolicy } from "./PermissionPolicy.ts";
import { ScriptedProvider } from "../providers/ScriptedProvider.ts";
import { ToolRegistry } from "./ToolRegistry.ts";

test("runtime executes an allowed tool call and finishes with an audited result", async () => {
  const registry = new ToolRegistry();
  registry.register({
    name: "crm.account.lookup",
    description: "Looks up an account",
    risk: "read",
    handler: async (_ctx, input) => ({
      content: { accountId: requireString(input.account_id), name: "Acme" },
    }),
  });

  const audit = new InMemoryAuditLog();
  const runtime = new AgentRuntime({
    provider: new ScriptedProvider([
      {
        type: "tool_call",
        toolName: "crm.account.lookup",
        input: { account_id: "acc_1" },
        reason: "Need the account before answering",
      },
      { type: "final", message: "Acme is ready." },
    ]),
    registry,
    policy: new AllowAllPolicy(),
    auditLog: audit,
  });

  const result = await runtime.run({
    tenantId: "tenant_1",
    actor: "agent:test",
    goal: "Check account readiness",
  });

  assert.equal(result.status, "completed");
  assert.equal(result.finalMessage, "Acme is ready.");
  assert.equal(result.toolResults.length, 1);
  assert.deepEqual(result.toolResults[0]?.content, {
    accountId: "acc_1",
    name: "Acme",
  });
  assert.deepEqual(
    audit.list().map((event) => event.type),
    [
      "run_started",
      "agent_decision",
      "permission_checked",
      "tool_started",
      "tool_finished",
      "agent_decision",
      "run_finished",
    ],
  );
});

test("runtime pauses before executing a tool that requires approval", async () => {
  const registry = new ToolRegistry();
  let called = false;
  registry.register({
    name: "shell.exec",
    description: "Runs a shell command",
    risk: "shell",
    handler: async () => {
      called = true;
      return { content: { ok: true } };
    },
  });

  const runtime = new AgentRuntime({
    provider: new ScriptedProvider([
      {
        type: "tool_call",
        toolName: "shell.exec",
        input: { command: "go test ./..." },
        reason: "Need verification",
      },
    ]),
    registry,
    policy: new StaticPolicy({ "shell.exec": "approval_required" }),
    auditLog: new InMemoryAuditLog(),
  });

  const result = await runtime.run({
    tenantId: "tenant_1",
    actor: "agent:test",
    goal: "Run verification",
  });

  assert.equal(result.status, "approval_required");
  assert.equal(result.pendingApproval?.toolName, "shell.exec");
  assert.equal(called, false);
});

test("runtime denies a forbidden tool without executing it", async () => {
  const registry = new ToolRegistry();
  let called = false;
  registry.register({
    name: "crm.account.delete",
    description: "Deletes an account",
    risk: "write",
    handler: async () => {
      called = true;
      return { content: { deleted: true } };
    },
  });

  const runtime = new AgentRuntime({
    provider: new ScriptedProvider([
      {
        type: "tool_call",
        toolName: "crm.account.delete",
        input: { account_id: "acc_1" },
      },
    ]),
    registry,
    policy: new StaticPolicy({ "crm.account.delete": "deny" }),
    auditLog: new InMemoryAuditLog(),
  });

  const result = await runtime.run({
    tenantId: "tenant_1",
    actor: "agent:test",
    goal: "Delete test account",
  });

  assert.equal(result.status, "denied");
  assert.equal(result.denial?.toolName, "crm.account.delete");
  assert.equal(called, false);
});

function requireString(value: unknown): string {
  if (typeof value !== "string") {
    throw new Error("expected string");
  }
  return value;
}

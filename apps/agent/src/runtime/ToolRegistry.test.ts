import test from "node:test";
import assert from "node:assert/strict";

import { ToolRegistry } from "./ToolRegistry.ts";

test("registry lists registered tools without exposing handlers", () => {
  const registry = new ToolRegistry();
  registry.register({
    name: "crm.accounts.list",
    description: "Lists accounts",
    risk: "read",
    handler: async () => ({ content: [] }),
  });

  assert.deepEqual(registry.describeTools(), [
    {
      name: "crm.accounts.list",
      description: "Lists accounts",
      risk: "read",
    },
  ]);
});

test("registry rejects duplicate tool names", () => {
  const registry = new ToolRegistry();
  const tool = {
    name: "crm.accounts.list",
    description: "Lists accounts",
    risk: "read" as const,
    handler: async () => ({ content: [] }),
  };

  registry.register(tool);

  assert.throws(() => registry.register(tool), /already registered/);
});

test("registry validates tool input before execution", async () => {
  const registry = new ToolRegistry();
  registry.register({
    name: "crm.accounts.create",
    description: "Creates an account",
    risk: "write",
    validateInput: (input) => {
      if (typeof input.name !== "string") {
        throw new Error("name is required");
      }
    },
    handler: async (_ctx, input) => ({
      content: {
        id: requireString(input.id),
        name: requireString(input.name),
      },
    }),
  });

  await assert.rejects(
    registry.execute(
      "crm.accounts.create",
      { tenantId: "tenant_1", actor: "agent:test" },
      { id: "acc_1" },
    ),
    /name is required/,
  );
});

function requireString(value: unknown): string {
  if (typeof value !== "string") {
    throw new Error("expected string");
  }
  return value;
}

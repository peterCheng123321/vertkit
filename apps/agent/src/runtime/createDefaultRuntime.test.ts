import test from "node:test";
import assert from "node:assert/strict";

import { createDefaultRuntime } from "./createDefaultRuntime.ts";

test("default runtime registers CRM tools and returns a clear message without provider config", async () => {
  const bundle = createDefaultRuntime({
    env: {
      VERTKIT_CRM_BASE_URL: "http://crm.local",
    },
  });

  const toolNames = bundle.registry.describeTools().map((tool) => tool.name);
  assert.ok(toolNames.includes("crm.accounts.list"));
  assert.ok(toolNames.includes("crm.rules.evaluate"));

  const result = await bundle.runtime.run({
    tenantId: "tenant_1",
    actor: "agent:test",
    goal: "Summarize accounts",
  });
  const secondResult = await bundle.runtime.run({
    tenantId: "tenant_1",
    actor: "agent:test",
    goal: "Summarize accounts again",
  });

  assert.equal(result.status, "completed");
  assert.match(result.finalMessage ?? "", /VERTKIT_AGENT_API_KEY/);
  assert.equal(secondResult.status, "completed");
  assert.equal(secondResult.finalMessage, result.finalMessage);
});

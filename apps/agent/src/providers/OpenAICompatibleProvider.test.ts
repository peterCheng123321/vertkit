import test from "node:test";
import assert from "node:assert/strict";

import { parseProviderDecision } from "./OpenAICompatibleProvider.ts";

test("provider parser accepts a JSON tool-call decision", () => {
  const decision = parseProviderDecision(
    JSON.stringify({
      type: "tool_call",
      toolName: "crm.accounts.list",
      input: {},
      reason: "Need current accounts",
    }),
  );

  assert.deepEqual(decision, {
    type: "tool_call",
    toolName: "crm.accounts.list",
    input: {},
    reason: "Need current accounts",
  });
});

test("provider parser accepts fenced JSON final decisions", () => {
  const decision = parseProviderDecision(
    '```json\n{"type":"final","message":"No changes needed."}\n```',
  );

  assert.deepEqual(decision, {
    type: "final",
    message: "No changes needed.",
  });
});

test("provider parser rejects malformed decisions", () => {
  assert.throws(
    () => parseProviderDecision('{"type":"tool_call","toolName":"crm.accounts.list"}'),
    /input must be a JSON object/,
  );
});

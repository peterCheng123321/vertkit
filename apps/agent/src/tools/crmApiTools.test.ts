import test from "node:test";
import assert from "node:assert/strict";

import { createCrmApiTools } from "./crmApiTools.ts";
import type { JsonObject } from "../types.ts";

test("crm account list tool calls the tenant-scoped backend route with service auth", async () => {
  const calls: Array<{ url: string; init: RequestInit }> = [];
  const tools = createCrmApiTools({
    baseUrl: "http://crm.local",
    serviceToken: "dev-token",
    fetchImpl: async (url, init) => {
      calls.push({ url: String(url), init: init ?? {} });
      return jsonResponse([{ id: "acc_1", name: "Acme" }]);
    },
  });

  const listAccounts = tools.find((tool) => tool.name === "crm.accounts.list");
  assert.ok(listAccounts);

  const result = await listAccounts.handler(
    { tenantId: "tenant_1", actor: "agent:test" },
    {},
  );

  assert.deepEqual(result.content, [{ id: "acc_1", name: "Acme" }]);
  assert.equal(calls[0]?.url, "http://crm.local/tenants/tenant_1/accounts");
  assert.equal((calls[0]?.init.headers as Record<string, string>).authorization, "Bearer dev-token");
});

test("crm account create tool validates required fields before calling backend", async () => {
  let called = false;
  const tools = createCrmApiTools({
    baseUrl: "http://crm.local",
    fetchImpl: async () => {
      called = true;
      return jsonResponse({ id: "acc_1" });
    },
  });

  const createAccount = tools.find((tool) => tool.name === "crm.accounts.create");
  assert.ok(createAccount);

  assert.throws(() => createAccount.validateInput?.({ id: "acc_1" }), /name is required/);
  assert.equal(called, false);
});

test("crm rule evaluation tool uses entity type in the route and sends the snapshot", async () => {
  const calls: Array<{ url: string; body: JsonObject }> = [];
  const tools = createCrmApiTools({
    baseUrl: "http://crm.local/",
    fetchImpl: async (url, init) => {
      calls.push({
        url: String(url),
        body: JSON.parse(String(init?.body)) as JsonObject,
      });
      return jsonResponse({ has_blocking: false, results: [] });
    },
  });

  const evaluateRules = tools.find((tool) => tool.name === "crm.rules.evaluate");
  assert.ok(evaluateRules);

  const result = await evaluateRules.handler(
    { tenantId: "tenant_1", actor: "agent:test" },
    {
      entity_type: "opportunity",
      payload: { opportunity: { id: "opp_1" }, operation: "update" },
    },
  );

  assert.deepEqual(result.content, { has_blocking: false, results: [] });
  assert.equal(
    calls[0]?.url,
    "http://crm.local/tenants/tenant_1/rules/evaluate/opportunity",
  );
  assert.deepEqual(calls[0]?.body, {
    opportunity: { id: "opp_1" },
    operation: "update",
  });
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

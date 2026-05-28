import type { JsonObject, JsonValue } from "../types.ts";
import { asJsonObject } from "../types.ts";
import type { AgentTool, ToolContext, ToolResult } from "../runtime/ToolRegistry.ts";

export interface CrmApiToolOptions {
  baseUrl: string;
  serviceToken?: string;
  fetchImpl?: typeof fetch;
}

export function createCrmApiTools(options: CrmApiToolOptions): AgentTool[] {
  const client = new CrmApiClient(options);
  return [
    {
      name: "crm.health",
      description: "Checks whether the VertKit CRM backend is reachable",
      risk: "read",
      handler: async () => client.get("/health"),
    },
    {
      name: "crm.tenants.create",
      description: "Creates a VertKit tenant",
      risk: "write",
      validateInput: (input) => {
        requireString(input, "id");
        requireString(input, "name");
        requireString(input, "default_currency");
        requireString(input, "default_locale");
      },
      handler: async (_ctx, input) => client.post("/tenants", input),
    },
    {
      name: "crm.accounts.list",
      description: "Lists accounts for the current tenant",
      risk: "read",
      handler: async (ctx) => client.get(tenantPath(ctx, "/accounts")),
    },
    {
      name: "crm.accounts.create",
      description: "Creates an account for the current tenant",
      risk: "write",
      validateInput: (input) => {
        requireString(input, "id");
        requireString(input, "name");
      },
      handler: async (ctx, input) => client.post(tenantPath(ctx, "/accounts"), input),
    },
    {
      name: "crm.contacts.list",
      description: "Lists contacts for the current tenant",
      risk: "read",
      handler: async (ctx) => client.get(tenantPath(ctx, "/contacts")),
    },
    {
      name: "crm.contacts.create",
      description: "Creates a contact for the current tenant",
      risk: "write",
      validateInput: (input) => {
        requireString(input, "id");
        requireString(input, "first_name");
        requireString(input, "last_name");
      },
      handler: async (ctx, input) => client.post(tenantPath(ctx, "/contacts"), input),
    },
    {
      name: "crm.opportunities.list",
      description: "Lists opportunities for the current tenant",
      risk: "read",
      handler: async (ctx) => client.get(tenantPath(ctx, "/opportunities")),
    },
    {
      name: "crm.opportunities.create",
      description: "Creates an opportunity for the current tenant",
      risk: "write",
      validateInput: (input) => {
        requireString(input, "id");
        requireString(input, "account_id");
        requireString(input, "name");
      },
      handler: async (ctx, input) => client.post(tenantPath(ctx, "/opportunities"), input),
    },
    {
      name: "crm.rules.evaluate",
      description: "Evaluates tenant business rules for an entity snapshot",
      risk: "read",
      validateInput: (input) => {
        requireString(input, "entity_type");
        asJsonObject(input.payload, "payload");
      },
      handler: async (ctx, input) => {
        const entityType = requireString(input, "entity_type");
        const payload = asJsonObject(input.payload, "payload");
        return client.post(tenantPath(ctx, `/rules/evaluate/${encodeURIComponent(entityType)}`), payload);
      },
    },
  ];
}

class CrmApiClient {
  private baseUrl: string;
  private serviceToken: string | undefined;
  private fetchImpl: typeof fetch;

  constructor(options: CrmApiToolOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.serviceToken = options.serviceToken;
    this.fetchImpl = options.fetchImpl ?? fetch;
  }

  async get(path: string): Promise<ToolResult> {
    return this.request("GET", path);
  }

  async post(path: string, body: JsonObject): Promise<ToolResult> {
    return this.request("POST", path, body);
  }

  private async request(method: string, path: string, body?: JsonObject): Promise<ToolResult> {
    const headers: Record<string, string> = {
      accept: "application/json",
    };
    if (body !== undefined) {
      headers["content-type"] = "application/json";
    }
    if (this.serviceToken !== undefined && this.serviceToken !== "") {
      headers.authorization = `Bearer ${this.serviceToken}`;
    }

    const requestInit: RequestInit = {
      method,
      headers,
    };
    if (body !== undefined) {
      requestInit.body = JSON.stringify(body);
    }

    const response = await this.fetchImpl(`${this.baseUrl}${path}`, requestInit);

    const text = await response.text();
    const content = text.trim() === "" ? null : (JSON.parse(text) as JsonValue);

    if (!response.ok) {
      throw new Error(`CRM API ${method} ${path} failed: ${response.status} ${text}`);
    }

    return { content };
  }
}

function tenantPath(ctx: ToolContext, suffix: string): string {
  return `/tenants/${encodeURIComponent(ctx.tenantId)}${suffix}`;
}

function requireString(input: JsonObject, field: string): string {
  const value = input[field];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${field} is required`);
  }
  return value;
}

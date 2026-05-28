import { InMemoryAuditLog } from "../audit/AuditLog.ts";
import { MissingProvider } from "../providers/MissingProvider.ts";
import { OpenAICompatibleProvider } from "../providers/OpenAICompatibleProvider.ts";
import type { AgentProvider } from "../providers/Provider.ts";
import { createCrmApiTools } from "../tools/crmApiTools.ts";
import { AgentRuntime } from "./AgentRuntime.ts";
import { RiskBasedPolicy } from "./PermissionPolicy.ts";
import { ToolRegistry } from "./ToolRegistry.ts";

export interface DefaultRuntimeEnv {
  [key: string]: string | undefined;
}

export interface DefaultRuntimeOptions {
  env?: DefaultRuntimeEnv;
}

export interface DefaultRuntimeBundle {
  runtime: AgentRuntime;
  registry: ToolRegistry;
  auditLog: InMemoryAuditLog;
}

export function createDefaultRuntime(options: DefaultRuntimeOptions = {}): DefaultRuntimeBundle {
  const env = options.env ?? process.env;
  const registry = new ToolRegistry();
  const crmBaseUrl = env.VERTKIT_CRM_BASE_URL ?? "http://127.0.0.1:8080";
  const serviceToken = env.VERTKIT_SERVICE_TOKEN;

  const crmOptions =
    serviceToken === undefined || serviceToken === ""
      ? { baseUrl: crmBaseUrl }
      : { baseUrl: crmBaseUrl, serviceToken };
  for (const tool of createCrmApiTools(crmOptions)) {
    registry.register(tool);
  }

  const auditLog = new InMemoryAuditLog();
  const provider = createProvider(env);
  const runtime = new AgentRuntime({
    provider,
    registry,
    policy: new RiskBasedPolicy(),
    auditLog,
  });

  return {
    runtime,
    registry,
    auditLog,
  };
}

function createProvider(env: DefaultRuntimeEnv): AgentProvider {
  const apiKey = env.VERTKIT_AGENT_API_KEY;
  const model = env.VERTKIT_AGENT_MODEL;
  if (apiKey !== undefined && apiKey !== "" && model !== undefined && model !== "") {
    return new OpenAICompatibleProvider({
      baseUrl: env.VERTKIT_AGENT_BASE_URL ?? "https://api.openai.com/v1",
      apiKey,
      model,
    });
  }

  return new MissingProvider(
    "No model provider configured. Set VERTKIT_AGENT_API_KEY, " +
      "VERTKIT_AGENT_MODEL, and optionally VERTKIT_AGENT_BASE_URL.",
  );
}

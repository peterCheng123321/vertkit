import type { AgentDecision, AgentProvider, AgentTurn } from "./Provider.ts";
import { asJsonObject } from "../types.ts";

export interface OpenAICompatibleProviderOptions {
  baseUrl: string;
  apiKey: string;
  model: string;
  fetchImpl?: typeof fetch;
}

export class OpenAICompatibleProvider implements AgentProvider {
  private baseUrl: string;
  private apiKey: string;
  private model: string;
  private fetchImpl: typeof fetch;

  constructor(options: OpenAICompatibleProviderOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.apiKey = options.apiKey;
    this.model = options.model;
    this.fetchImpl = options.fetchImpl ?? fetch;
  }

  async next(turn: AgentTurn): Promise<AgentDecision> {
    const response = await this.fetchImpl(`${this.baseUrl}/chat/completions`, {
      method: "POST",
      headers: {
        "content-type": "application/json",
        authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify({
        model: this.model,
        temperature: 0,
        messages: [
          {
            role: "system",
            content:
              "You are the VertKit agent runtime. Return only JSON. " +
              'Use {"type":"tool_call","toolName":"...","input":{}} to call a tool, ' +
              'or {"type":"final","message":"..."} when complete.',
          },
          {
            role: "user",
            content: JSON.stringify({
              tenant_id: turn.tenantId,
              actor: turn.actor,
              goal: turn.goal,
              step: turn.step,
              tools: turn.tools,
              tool_results: turn.toolResults,
            }),
          },
        ],
      }),
    });

    if (!response.ok) {
      throw new Error(`model provider failed: ${response.status} ${await response.text()}`);
    }

    const body = (await response.json()) as {
      choices?: Array<{ message?: { content?: string } }>;
    };
    const content = body.choices?.[0]?.message?.content;
    if (typeof content !== "string" || content.trim() === "") {
      throw new Error("model provider returned no message content");
    }
    return parseProviderDecision(content);
  }
}

export function parseProviderDecision(raw: string): AgentDecision {
  const parsed = asJsonObject(JSON.parse(stripJsonFence(raw)), "provider decision");
  if (parsed.type === "final") {
    if (typeof parsed.message !== "string" || parsed.message.trim() === "") {
      throw new Error("final decision message is required");
    }
    return {
      type: "final",
      message: parsed.message,
    };
  }

  if (parsed.type === "tool_call") {
    const toolName = parsed.toolName ?? parsed.tool_name;
    if (typeof toolName !== "string" || toolName.trim() === "") {
      throw new Error("tool_call decision toolName is required");
    }
    const decision: AgentDecision = {
      type: "tool_call",
      toolName,
      input: asJsonObject(parsed.input, "input"),
    };
    if (typeof parsed.reason === "string") {
      decision.reason = parsed.reason;
    }
    return decision;
  }

  throw new Error("provider decision type must be final or tool_call");
}

function stripJsonFence(raw: string): string {
  const trimmed = raw.trim();
  const match = /^```(?:json)?\s*([\s\S]*?)\s*```$/i.exec(trimmed);
  return match?.[1]?.trim() ?? trimmed;
}

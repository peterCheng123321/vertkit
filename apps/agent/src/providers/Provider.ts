import type { AuditEvent } from "../audit/AuditLog.ts";
import type { JsonObject } from "../types.ts";
import type { ToolDescription, ToolResult } from "../runtime/ToolRegistry.ts";

export type AgentDecision =
  | {
      type: "tool_call";
      toolName: string;
      input: JsonObject;
      reason?: string;
    }
  | {
      type: "final";
      message: string;
    };

export interface AgentTurn {
  runId: string;
  tenantId: string;
  actor: string;
  goal: string;
  step: number;
  tools: ToolDescription[];
  toolResults: ToolResult[];
  auditEvents: AuditEvent[];
}

export interface AgentProvider {
  next(turn: AgentTurn): Promise<AgentDecision>;
}

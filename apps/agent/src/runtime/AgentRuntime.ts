import type { AuditLog } from "../audit/AuditLog.ts";
import type { AgentDecision, AgentProvider } from "../providers/Provider.ts";
import type { JsonObject, JsonValue } from "../types.ts";
import { asJsonObject } from "../types.ts";
import type { PermissionDecision, PermissionPolicy } from "./PermissionPolicy.ts";
import type { ToolRegistry, ToolResult } from "./ToolRegistry.ts";

export type AgentRunStatus = "completed" | "approval_required" | "denied" | "failed";

export interface AgentRunInput {
  tenantId: string;
  actor: string;
  goal: string;
  maxSteps?: number;
}

export interface PendingApproval {
  toolName: string;
  input: JsonObject;
  reason?: string;
}

export interface Denial {
  toolName: string;
  input: JsonObject;
}

export interface AgentRunResult {
  runId: string;
  tenantId: string;
  actor: string;
  goal: string;
  status: AgentRunStatus;
  finalMessage?: string;
  toolResults: ToolResult[];
  pendingApproval?: PendingApproval;
  denial?: Denial;
  error?: string;
}

export interface AgentRuntimeOptions {
  provider: AgentProvider;
  registry: ToolRegistry;
  policy: PermissionPolicy;
  auditLog: AuditLog;
}

export class AgentRuntime {
  private provider: AgentProvider;
  private registry: ToolRegistry;
  private policy: PermissionPolicy;
  private auditLog: AuditLog;
  private nextRunId = 1;

  constructor(options: AgentRuntimeOptions) {
    this.provider = options.provider;
    this.registry = options.registry;
    this.policy = options.policy;
    this.auditLog = options.auditLog;
  }

  async run(input: AgentRunInput): Promise<AgentRunResult> {
    const runId = `run_${this.nextRunId++}`;
    const maxSteps = input.maxSteps ?? 8;
    const toolResults: ToolResult[] = [];

    await this.audit("run_started", runId, input, {
      goal: input.goal,
      max_steps: maxSteps,
    });

    try {
      for (let step = 0; step < maxSteps; step += 1) {
        const decision = await this.provider.next({
          runId,
          tenantId: input.tenantId,
          actor: input.actor,
          goal: input.goal,
          step,
          tools: this.registry.describeTools(),
          toolResults: structuredClone(toolResults),
          auditEvents: this.auditLog.list(runId),
        });

        await this.auditDecision(runId, input, decision);

        if (decision.type === "final") {
          const result: AgentRunResult = {
            runId,
            tenantId: input.tenantId,
            actor: input.actor,
            goal: input.goal,
            status: "completed",
            finalMessage: decision.message,
            toolResults,
          };
          await this.finish(runId, input, result);
          return result;
        }

        const maybeStopped = await this.handleToolCall(runId, input, decision, toolResults);
        if (maybeStopped !== undefined) {
          await this.finish(runId, input, maybeStopped);
          return maybeStopped;
        }
      }

      const result: AgentRunResult = {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "failed",
        error: `Agent exceeded max steps: ${maxSteps}`,
        toolResults,
      };
      await this.finish(runId, input, result);
      return result;
    } catch (error) {
      const result: AgentRunResult = {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "failed",
        error: error instanceof Error ? error.message : String(error),
        toolResults,
      };
      await this.finish(runId, input, result);
      return result;
    }
  }

  private async handleToolCall(
    runId: string,
    input: AgentRunInput,
    decision: Extract<AgentDecision, { type: "tool_call" }>,
    toolResults: ToolResult[],
  ): Promise<AgentRunResult | undefined> {
    const tool = this.registry.get(decision.toolName);
    if (tool === undefined) {
      return {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "failed",
        error: `Unknown tool: ${decision.toolName}`,
        toolResults,
      };
    }

    const toolInput = asJsonObject(decision.input, "tool input");
    const permission = await this.policy.check({
      tenantId: input.tenantId,
      actor: input.actor,
      toolName: decision.toolName,
      risk: tool.risk,
      input: toolInput,
    });
    await this.auditPermission(runId, input, decision.toolName, permission);

    if (permission === "approval_required") {
      const pendingApproval: PendingApproval = {
        toolName: decision.toolName,
        input: toolInput,
      };
      if (decision.reason !== undefined) {
        pendingApproval.reason = decision.reason;
      }
      return {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "approval_required",
        pendingApproval,
        toolResults,
      };
    }

    if (permission === "deny") {
      return {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "denied",
        denial: {
          toolName: decision.toolName,
          input: toolInput,
        },
        toolResults,
      };
    }

    await this.audit("tool_started", runId, input, {
      tool_name: decision.toolName,
    });
    try {
      const result = await this.registry.execute(
        decision.toolName,
        {
          tenantId: input.tenantId,
          actor: input.actor,
          runId,
        },
        toolInput,
      );
      toolResults.push(result);
      await this.audit("tool_finished", runId, input, {
        tool_name: decision.toolName,
        result: result.content,
      });
      return undefined;
    } catch (error) {
      await this.audit("tool_failed", runId, input, {
        tool_name: decision.toolName,
        error: error instanceof Error ? error.message : String(error),
      });
      return {
        runId,
        tenantId: input.tenantId,
        actor: input.actor,
        goal: input.goal,
        status: "failed",
        error: error instanceof Error ? error.message : String(error),
        toolResults,
      };
    }
  }

  private async finish(runId: string, input: AgentRunInput, result: AgentRunResult): Promise<void> {
    await this.audit("run_finished", runId, input, {
      status: result.status,
      final_message: result.finalMessage ?? null,
      error: result.error ?? null,
    });
  }

  private async auditDecision(runId: string, input: AgentRunInput, decision: AgentDecision): Promise<void> {
    await this.audit("agent_decision", runId, input, {
      decision_type: decision.type,
      tool_name: decision.type === "tool_call" ? decision.toolName : null,
      reason: decision.type === "tool_call" ? decision.reason ?? null : null,
    });
  }

  private async auditPermission(
    runId: string,
    input: AgentRunInput,
    toolName: string,
    decision: PermissionDecision,
  ): Promise<void> {
    await this.audit("permission_checked", runId, input, {
      tool_name: toolName,
      decision,
    });
  }

  private async audit(
    type: Parameters<AuditLog["append"]>[0]["type"],
    runId: string,
    input: AgentRunInput,
    data: Record<string, JsonValue>,
  ): Promise<void> {
    await this.auditLog.append({
      type,
      runId,
      tenantId: input.tenantId,
      actor: input.actor,
      data,
    });
  }
}

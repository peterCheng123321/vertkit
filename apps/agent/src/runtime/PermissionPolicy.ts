import type { JsonObject } from "../types.ts";
import type { ToolRisk } from "./ToolRegistry.ts";

export type PermissionDecision = "allow" | "approval_required" | "deny";

export interface PermissionRequest {
  tenantId: string;
  actor: string;
  toolName: string;
  risk: ToolRisk;
  input: JsonObject;
}

export interface PermissionPolicy {
  check(request: PermissionRequest): Promise<PermissionDecision>;
}

export class AllowAllPolicy implements PermissionPolicy {
  async check(): Promise<PermissionDecision> {
    return "allow";
  }
}

export class StaticPolicy implements PermissionPolicy {
  private decisions: Record<string, PermissionDecision>;
  private defaultDecision: PermissionDecision;

  constructor(decisions: Record<string, PermissionDecision>, defaultDecision: PermissionDecision = "allow") {
    this.decisions = decisions;
    this.defaultDecision = defaultDecision;
  }

  async check(request: PermissionRequest): Promise<PermissionDecision> {
    return this.decisions[request.toolName] ?? this.defaultDecision;
  }
}

export class RiskBasedPolicy implements PermissionPolicy {
  async check(request: PermissionRequest): Promise<PermissionDecision> {
    if (request.risk === "shell" || request.risk === "external") {
      return "approval_required";
    }
    if (request.risk === "write") {
      return "approval_required";
    }
    return "allow";
  }
}

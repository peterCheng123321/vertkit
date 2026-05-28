import type { JsonObject } from "../types.ts";

export type AuditEventType =
  | "run_started"
  | "agent_decision"
  | "permission_checked"
  | "tool_started"
  | "tool_finished"
  | "tool_failed"
  | "run_finished";

export interface AuditEvent {
  id: string;
  runId: string;
  tenantId: string;
  actor: string;
  type: AuditEventType;
  timestamp: string;
  data: JsonObject;
}

export interface AuditLog {
  append(event: Omit<AuditEvent, "id" | "timestamp">): Promise<AuditEvent>;
  list(runId?: string): AuditEvent[];
}

export class InMemoryAuditLog implements AuditLog {
  private events: AuditEvent[] = [];
  private nextId = 1;

  async append(event: Omit<AuditEvent, "id" | "timestamp">): Promise<AuditEvent> {
    const recorded: AuditEvent = {
      ...event,
      id: `evt_${this.nextId++}`,
      timestamp: new Date().toISOString(),
    };
    this.events.push(recorded);
    return recorded;
  }

  list(runId?: string): AuditEvent[] {
    const events = runId === undefined ? this.events : this.events.filter((event) => event.runId === runId);
    return events.map((event) => ({
      ...event,
      data: structuredClone(event.data),
    }));
  }
}

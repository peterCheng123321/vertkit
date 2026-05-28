import type { JsonObject, JsonValue } from "../types.ts";

export type ToolRisk = "read" | "write" | "external" | "shell";

export interface ToolContext {
  tenantId: string;
  actor: string;
  runId?: string;
}

export interface ToolResult {
  content: JsonValue;
  metadata?: JsonObject;
}

export interface AgentTool {
  name: string;
  description: string;
  risk: ToolRisk;
  validateInput?: (input: JsonObject) => void;
  handler: (context: ToolContext, input: JsonObject) => Promise<ToolResult>;
}

export interface ToolDescription {
  name: string;
  description: string;
  risk: ToolRisk;
}

export class ToolRegistry {
  private tools = new Map<string, AgentTool>();

  register(tool: AgentTool): void {
    if (this.tools.has(tool.name)) {
      throw new Error(`Tool ${tool.name} is already registered`);
    }
    this.tools.set(tool.name, tool);
  }

  get(name: string): AgentTool | undefined {
    return this.tools.get(name);
  }

  describeTools(): ToolDescription[] {
    return Array.from(this.tools.values()).map((tool) => ({
      name: tool.name,
      description: tool.description,
      risk: tool.risk,
    }));
  }

  async execute(name: string, context: ToolContext, input: JsonObject): Promise<ToolResult> {
    const tool = this.tools.get(name);
    if (tool === undefined) {
      throw new Error(`Unknown tool: ${name}`);
    }
    tool.validateInput?.(input);
    return tool.handler(context, input);
  }
}

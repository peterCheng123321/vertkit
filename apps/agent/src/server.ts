import http, { type IncomingMessage, type ServerResponse } from "node:http";
import type { AddressInfo } from "node:net";

import type { AgentRuntime } from "./runtime/AgentRuntime.ts";
import type { ToolRegistry } from "./runtime/ToolRegistry.ts";
import { asJsonObject } from "./types.ts";

export interface AgentServerOptions {
  runtime: AgentRuntime;
  registry: ToolRegistry;
}

export interface AgentServer {
  listen(port: number): Promise<void>;
  close(): Promise<void>;
  port(): number;
}

export function createAgentServer(options: AgentServerOptions): AgentServer {
  const server = http.createServer(async (request, response) => {
    try {
      await route(request, response, options);
    } catch (error) {
      writeJson(response, 500, {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  });

  return {
    listen(port: number): Promise<void> {
      return new Promise((resolve, reject) => {
        server.once("error", reject);
        server.listen(port, "127.0.0.1", () => {
          server.off("error", reject);
          resolve();
        });
      });
    },
    close(): Promise<void> {
      return new Promise((resolve, reject) => {
        server.close((error) => {
          if (error !== undefined) {
            reject(error);
            return;
          }
          resolve();
        });
      });
    },
    port(): number {
      const address = server.address() as AddressInfo | null;
      if (address === null) {
        throw new Error("Server is not listening");
      }
      return address.port;
    },
  };
}

async function route(
  request: IncomingMessage,
  response: ServerResponse,
  options: AgentServerOptions,
): Promise<void> {
  if (request.method === "GET" && request.url === "/health") {
    writeJson(response, 200, {
      ok: true,
      service: "vertkit-agent",
    });
    return;
  }

  if (request.method === "GET" && request.url === "/agent/tools") {
    writeJson(response, 200, {
      tools: options.registry.describeTools(),
    });
    return;
  }

  if (request.method === "POST" && request.url === "/agent/jobs") {
    const body = asJsonObject(await readJson(request), "request body");
    const tenantId = requireString(body.tenant_id, "tenant_id");
    const goal = requireString(body.goal, "goal");
    const actor = typeof body.actor === "string" ? body.actor : "agent:http";
    const maxSteps = typeof body.max_steps === "number" ? body.max_steps : undefined;

    const runInput = {
      tenantId,
      actor,
      goal,
    };
    const result = await options.runtime.run(
      maxSteps === undefined ? runInput : { ...runInput, maxSteps },
    );

    writeJson(response, 200, {
      run_id: result.runId,
      tenant_id: result.tenantId,
      actor: result.actor,
      goal: result.goal,
      status: result.status,
      final_message: result.finalMessage ?? null,
      tool_results: result.toolResults,
      pending_approval: result.pendingApproval ?? null,
      denial: result.denial ?? null,
      error: result.error ?? null,
    });
    return;
  }

  writeJson(response, 404, {
    error: "not found",
  });
}

async function readJson(request: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = [];
  for await (const chunk of request) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  if (chunks.length === 0) {
    return {};
  }
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

function requireString(value: unknown, name: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${name} is required`);
  }
  return value;
}

function writeJson(response: ServerResponse, statusCode: number, body: unknown): void {
  response.writeHead(statusCode, {
    "content-type": "application/json",
  });
  response.end(JSON.stringify(body));
}

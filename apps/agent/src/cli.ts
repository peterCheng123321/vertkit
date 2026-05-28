#!/usr/bin/env node
import { parseCliArgs } from "./cliArgs.ts";
import { createDefaultRuntime } from "./runtime/createDefaultRuntime.ts";
import { createAgentServer } from "./server.ts";

async function main(): Promise<void> {
  const command = parseCliArgs(process.argv.slice(2));
  const bundle = createDefaultRuntime();

  if (command.command === "help") {
    printHelp();
    return;
  }

  if (command.command === "tools") {
    console.log(JSON.stringify({ tools: bundle.registry.describeTools() }, null, 2));
    return;
  }

  if (command.command === "run") {
    const result = await bundle.runtime.run({
      tenantId: process.env.VERTKIT_TENANT_ID ?? "default",
      actor: process.env.VERTKIT_AGENT_ACTOR ?? "agent:cli",
      goal: command.goal,
    });
    console.log(JSON.stringify(result, null, 2));
    return;
  }

  const server = createAgentServer({
    runtime: bundle.runtime,
    registry: bundle.registry,
  });
  await server.listen(command.port);
  console.log(`vertkit-agent listening on http://127.0.0.1:${server.port()}`);
}

function printHelp(): void {
  console.log(`vertkit-agent

Usage:
  npm run agent:start -- serve [--port 8787]
  npm run agent:start -- tools
  npm run agent:start -- run <goal>

Environment:
  VERTKIT_CRM_BASE_URL       Go CRM backend URL, default http://127.0.0.1:8080
  VERTKIT_SERVICE_TOKEN      service token for tenant-scoped CRM calls
  VERTKIT_AGENT_BASE_URL     OpenAI-compatible chat completions URL
  VERTKIT_AGENT_API_KEY      model provider API key
  VERTKIT_AGENT_MODEL        model name
  VERTKIT_TENANT_ID          tenant for CLI run command
`);
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exitCode = 1;
});

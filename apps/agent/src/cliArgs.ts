export type CliCommand =
  | {
      command: "serve";
      port: number;
    }
  | {
      command: "run";
      goal: string;
    }
  | {
      command: "tools";
    }
  | {
      command: "help";
    };

export function parseCliArgs(args: string[]): CliCommand {
  const [command = "serve", ...rest] = args;

  if (command === "serve") {
    return {
      command: "serve",
      port: parsePort(rest),
    };
  }

  if (command === "run") {
    const goal = rest.join(" ").trim();
    if (goal === "") {
      throw new Error("goal is required");
    }
    return {
      command: "run",
      goal,
    };
  }

  if (command === "tools") {
    return { command: "tools" };
  }

  if (command === "help" || command === "--help" || command === "-h") {
    return { command: "help" };
  }

  throw new Error(`unknown command: ${command}`);
}

function parsePort(args: string[]): number {
  const index = args.indexOf("--port");
  if (index === -1) {
    return 8787;
  }
  const raw = args[index + 1];
  const port = Number(raw);
  if (!Number.isInteger(port) || port <= 0 || port > 65535) {
    throw new Error("--port must be a valid TCP port");
  }
  return port;
}

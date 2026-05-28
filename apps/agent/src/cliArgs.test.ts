import test from "node:test";
import assert from "node:assert/strict";

import { parseCliArgs } from "./cliArgs.ts";

test("cli parser defaults to serve on port 8787", () => {
  assert.deepEqual(parseCliArgs([]), {
    command: "serve",
    port: 8787,
  });
});

test("cli parser accepts a run goal", () => {
  assert.deepEqual(parseCliArgs(["run", "Build", "healthcare", "CRM"]), {
    command: "run",
    goal: "Build healthcare CRM",
  });
});

test("cli parser accepts explicit serve port", () => {
  assert.deepEqual(parseCliArgs(["serve", "--port", "9001"]), {
    command: "serve",
    port: 9001,
  });
});

test("cli parser rejects run without a goal", () => {
  assert.throws(() => parseCliArgs(["run"]), /goal is required/);
});

import type { AgentDecision, AgentProvider, AgentTurn } from "./Provider.ts";

export class ScriptedProvider implements AgentProvider {
  private decisions: AgentDecision[];
  private index = 0;

  constructor(decisions: AgentDecision[]) {
    this.decisions = decisions;
  }

  async next(_turn: AgentTurn): Promise<AgentDecision> {
    const decision = this.decisions[this.index];
    this.index += 1;
    if (decision === undefined) {
      return {
        type: "final",
        message: "No further scripted decisions.",
      };
    }
    return decision;
  }
}

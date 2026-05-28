import type { AgentDecision, AgentProvider, AgentTurn } from "./Provider.ts";

export class MissingProvider implements AgentProvider {
  private message: string;

  constructor(message: string) {
    this.message = message;
  }

  async next(_turn: AgentTurn): Promise<AgentDecision> {
    return {
      type: "final",
      message: this.message,
    };
  }
}

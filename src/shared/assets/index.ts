import agentClaude from "./agent-claude.png";
import agentCodex from "./agent-codex.png";
import agentCursor from "./agent-cursor.png";
import agentGemini from "./agent-gemini.png";
import backdrop from "./backdrop.jpg";
import mark from "./mark.png";

export { backdrop, mark };

/** The logos of coding agents, by skill target. */
export const agentLogos: Record<string, string | undefined> = {
  claude: agentClaude,
  codex: agentCodex,
  cursor: agentCursor,
  gemini: agentGemini,
};

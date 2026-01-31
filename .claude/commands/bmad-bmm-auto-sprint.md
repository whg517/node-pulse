---
name: 'auto-sprint'
description: 'Automatic sprint orchestration - continuously executes stories from backlog through done with intelligent workflow routing'
---

IT IS CRITICAL THAT YOU FOLLOW THESE STEPS:

<steps CRITICAL="TRUE">
1. Launch the BMAD Auto-Sprint Agent using the agent activation command
2. The agent will orchestrate all sprint execution using BMAD sub-agents
3. The agent runs continuously in a dedicated background context
4. You can resume at any time by running this skill again
</steps>

<execution>
The auto-sprint agent will:
- Use Skill tool to invoke BMAD workflows with proper agents
- Execute continuously in dedicated agent context
- Coordinate bmad-agent-bmm-sm for create-story and retrospective
- Coordinate bmad-agent-bmm-dev for dev-story and code-review
- Handle git commits after review completion
- Continue iterating until all stories complete or you halt it
- Allow intervention and resumption at any time
</execution>

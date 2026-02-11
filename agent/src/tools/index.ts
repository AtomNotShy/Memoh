import { AuthFetcher } from '..'
import { AgentAction, AgentAuthContext, BraveConfig, IdentityContext, ModelConfig } from '../types'
import { ToolSet } from 'ai'
import { getWebTools } from './web'
import { getScheduleTools } from './schedule'
import { getMemoryTools } from './memory'
import { getSubagentTools } from './subagent'
import { getContactTools } from './contact'
import { getMessageTools } from './message'
import { getSkillTools } from './skill'

export interface ToolsParams {
  fetch: AuthFetcher
  model: ModelConfig
  brave?: BraveConfig
  identity: IdentityContext
  auth: AgentAuthContext
  enableSkill: (skill: string) => void
}

export const getTools = (
  actions: AgentAction[],
  { fetch, model, brave, identity, auth, enableSkill }: ToolsParams
) => {
  const tools: ToolSet = {}
  if (actions.includes(AgentAction.Web) && brave) {
    const webTools = getWebTools({ brave })
    Object.assign(tools, webTools)
  }
  if (actions.includes(AgentAction.Schedule)) {
    const scheduleTools = getScheduleTools({ fetch, identity })
    Object.assign(tools, scheduleTools)
  }
  if (actions.includes(AgentAction.Memory)) {
    const memoryTools = getMemoryTools({ fetch, identity })
    Object.assign(tools, memoryTools)
  }
  if (actions.includes(AgentAction.Subagent)) {
    const subagentTools = getSubagentTools({ fetch, model, brave, identity, auth })
    Object.assign(tools, subagentTools)
  }
  if (actions.includes(AgentAction.Contact)) {
    const contactTools = getContactTools({ fetch, identity })
    Object.assign(tools, contactTools)
  }
  if (actions.includes(AgentAction.Message)) {
    const messageTools = getMessageTools({ fetch, identity })
    Object.assign(tools, messageTools)
  }
  if (actions.includes(AgentAction.Skill)) {
    const skillTools = getSkillTools({ useSkill: enableSkill })
    Object.assign(tools, skillTools)
  }
    return tools
}
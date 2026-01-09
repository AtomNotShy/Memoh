import { Message } from 'xsai'
import { MemoryUnit } from './memory-unit'

export const rawMessages = (messages: Message[]) => {
  return messages.map((message) => {
    if (message.role === 'user') {
      return `User: ${message.content}`
    } else if (message.role === 'assistant') {
      let toolCalls = ''
      if (message.tool_calls && message.tool_calls.length !== 0) {
        toolCalls = `Tool Calls: ${message.tool_calls.map(t => t.function.name).join(', ')}`
      }
      return `You: ${message.content} \n${toolCalls}`
    } else if (message.role === 'tool') {
      return `Tool Result: ${message.content}`
    } else {
      return null
    }
  })
  .filter((message) => message !== null)
  .join('\n\n')
}

export const rawMemory = (memory: MemoryUnit, locale: Intl.LocalesArgument) => {
  return `
  ---
  date: ${memory.timestamp.toLocaleDateString(locale)}
  time: ${memory.timestamp.toLocaleTimeString(locale)}
  timezone: ${memory.timestamp.getTimezoneOffset()}
  ---
  ${rawMessages(memory.messages)}
  `.trim()
}

import { embed } from 'xsai'
import { EmbedParams } from './types'
import { MemoryUnit } from './memory-unit'
import { rawMemory } from './raw'
import { db } from '@byteagent/db'
import { memory } from '@byteagent/db/schema'

export interface AddMemoryParams extends EmbedParams {
  locale: Intl.LocalesArgument
}

export interface AddMemoryInput {
  memory: MemoryUnit
}

export const createAddMemory = (params: AddMemoryParams) =>
  async ({ memory: memoryUnit }: AddMemoryInput) => {
    const rawContent = rawMemory(memoryUnit, params.locale)
    const { embedding } = await embed({
      model: params.model,
      input: rawContent,
      apiKey: params.apiKey,
      baseURL: params.baseURL,
    })
    await db.insert(memory)
      .values({
        timestamp: memoryUnit.timestamp,
        user: memoryUnit.user,
        rawContent,
        embedding,
        messages: memoryUnit.messages,
      })
      .onConflictDoNothing()
  }
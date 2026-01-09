import { embed } from 'xsai'
import { filterByEmbedding } from './filter'
import { EmbedParams } from './types'

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface MemorySearchParams extends EmbedParams { }

export interface MemorySearchInput {
  user: string
  query: string
  maxResults?: number
}

export const createMemorySearch = (params: MemorySearchParams) =>
  async ({ user, query, maxResults = 10 }: MemorySearchInput) => {
    const { embedding } = await embed({
      model: params.model,
      input: query,
      apiKey: params.apiKey,
      baseURL: params.baseURL,
    })
    return await filterByEmbedding(embedding, user, maxResults)
  }
import { db } from '@byteagent/db'
import { memory } from '@byteagent/db/schema'
import { and, gte, lte, asc, sql, cosineDistance, gt, desc, eq } from 'drizzle-orm'
import { MemoryUnit } from './memory-unit'

export const filterByTimestamp = async (
  from: Date,
  to: Date,
  user: string,
) => {
  const results = await db
    .select()
    .from(memory)
    .where(and(
      gte(memory.timestamp, from),
      lte(memory.timestamp, to),
      eq(memory.user, user),
    ))
    .orderBy(asc(memory.timestamp))

  return results.map((result) => ({
    messages: result.messages,
    timestamp: new Date(result.timestamp),
    user: result.user,
    raw: result.rawContent,
  })) as MemoryUnit[]
}

export const filterByEmbedding = async (
  embedding: number[],
  user: string,
  limit: number = 10,
) => {
  const similarity = sql<number>`1 - (${cosineDistance(memory.embedding, embedding)})`
  const results = await db
    .select({
      similarity,
      messages: memory.messages,
      timestamp: memory.timestamp,
      user: memory.user,
      rawContent: memory.rawContent,
      embedding: memory.embedding,
      id: memory.id,
    })
    .from(memory)
    .where(and(
      gt(similarity, 0.5),
      eq(memory.user, user),
    ))
    .orderBy((t) => desc(t.similarity))
    .limit(limit)
  return results.map((result) => ({
    messages: result.messages,
    timestamp: new Date(result.timestamp),
    user: result.user,
    raw: result.rawContent,
  })) as MemoryUnit[]
}
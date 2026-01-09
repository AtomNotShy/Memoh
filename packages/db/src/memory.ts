import { pgTable, timestamp, uuid, jsonb, text, vector, index } from 'drizzle-orm/pg-core'

export const memory = pgTable(
  'memory', 
  {
    id: uuid('id').primaryKey().defaultRandom(),
    messages: jsonb('messages').notNull(),
    timestamp: timestamp('timestamp').notNull(),
    user: text('user').notNull(),
    rawContent: text('raw_content').notNull(),
    embedding: vector('embedding', { dimensions: 1536 }).notNull(),
  },
  (table) => [
    index('embedding_index').using('hnsw', table.embedding.op('vector_cosine_ops')),
  ]
)
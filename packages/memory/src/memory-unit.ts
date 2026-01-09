import { Message } from 'xsai'

export interface MemoryUnit {
  messages: Message[]
  timestamp: Date
  user: string
  raw: string
}
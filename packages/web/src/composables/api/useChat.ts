import { fetchApi } from '@/utils/request'

export interface Bot {
  id: string
  display_name?: string
  avatar_url?: string
  type?: string
}

export interface BotsResponse {
  items: Bot[]
}

export interface ChatSummary {
  id: string
  bot_id: string
  kind: string
  title?: string
  created_at?: string
  updated_at?: string
  access_mode?: 'participant' | 'channel_identity_observed'
  participant_role?: string
  last_observed_at?: string
}

export interface ChatsResponse {
  items: ChatSummary[]
}

export interface ModelMessage {
  role: string
  content?: unknown
}

export interface ChatResponse {
  messages: ModelMessage[]
  skills?: string[]
  model?: string
  provider?: string
}

export interface PersistedChatMessage {
  id: string
  chat_id: string
  bot_id: string
  role: string
  content?: unknown
  created_at?: string
}

export interface ChatMessagesResponse {
  items: PersistedChatMessage[]
}

export async function fetchBots(): Promise<Bot[]> {
  const res = await fetchApi<BotsResponse>('/bots')
  return res.items
}

export async function fetchChats(botId: string): Promise<ChatSummary[]> {
  const res = await fetchApi<ChatsResponse>(`/bots/${botId}/chats`)
  return res.items ?? []
}

export async function createChat(botId: string): Promise<ChatSummary> {
  return fetchApi<ChatSummary>(`/bots/${botId}/chats`, {
    method: 'POST',
    body: {
      kind: 'direct',
    },
  })
}

export async function deleteChat(chatId: string): Promise<void> {
  await fetchApi(`/chats/${chatId}`, {
    method: 'DELETE',
  })
}

export async function resolveOrCreateChat(botId: string): Promise<string> {
  const chats = await fetchChats(botId)
  if (chats.length > 0 && chats[0]?.id) {
    return chats[0].id
  }

  const created = await createChat(botId)
  return created.id
}

export async function sendChatMessage(chatId: string, text: string): Promise<ChatResponse> {
  return fetchApi<ChatResponse>(`/chats/${chatId}/messages`, {
    method: 'POST',
    body: {
      query: text,
      current_channel: 'web',
      channels: ['web'],
    },
  })
}

export async function fetchChatMessages(chatId: string): Promise<PersistedChatMessage[]> {
  const res = await fetchApi<ChatMessagesResponse>(`/chats/${chatId}/messages`)
  return res.items ?? []
}

export async function streamChatMessage(
  chatId: string,
  text: string,
  onTextDelta: (delta: string) => void,
): Promise<ChatResponse | null> {
  const token = localStorage.getItem('token') ?? ''
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const response = await fetch(`/api/chats/${chatId}/messages/stream`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      query: text,
      current_channel: 'web',
      channels: ['web'],
    }),
  })

  if (!response.ok || !response.body) {
    const message = await response.text().catch(() => '')
    throw new Error(message || `Stream request failed: ${response.status}`)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let finalResponse: ChatResponse | null = null
  const handlePayload = (payload: string) => {
    if (!payload || payload === '[DONE]') {
      return
    }

    const event = parseStreamPayload(payload)
    if (typeof event === 'string') {
      if (event) {
        onTextDelta(event)
      }
      return
    }
    if (!event) {
      return
    }

    if (typeof event.error === 'string' && event.error.trim()) {
      throw new Error(event.error)
    }

    const eventType = String(event.type ?? '').toLowerCase()
    if (eventType === 'error') {
      const message = typeof event.message === 'string'
        ? event.message
        : typeof event.error === 'string'
          ? event.error
          : 'Stream error'
      throw new Error(message)
    }
    if (eventType === 'text_delta' && typeof event.delta === 'string') {
      onTextDelta(event.delta)
      return
    }
    if (eventType === 'agent_end' && Array.isArray(event.messages)) {
      finalResponse = {
        messages: event.messages as ModelMessage[],
        skills: Array.isArray(event.skills) ? event.skills.filter((item): item is string => typeof item === 'string') : undefined,
        model: typeof event.model === 'string' ? event.model : undefined,
        provider: typeof event.provider === 'string' ? event.provider : undefined,
      }
    }
  }

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }
    buffer += decoder.decode(value, { stream: true })

    let index = buffer.indexOf('\n')
    while (index >= 0) {
      const line = buffer.slice(0, index).trim()
      buffer = buffer.slice(index + 1)
      index = buffer.indexOf('\n')

      if (!line.startsWith('data:')) {
        continue
      }
      const payload = line.slice(5).trim()
      handlePayload(payload)
    }
  }

  const tail = buffer.trim()
  if (tail.startsWith('data:')) {
    handlePayload(tail.slice(5).trim())
  }

  return finalResponse
}

export function extractAssistantTexts(messages: ModelMessage[]): string[] {
  if (!Array.isArray(messages)) {
    return []
  }

  const outputs: string[] = []
  for (const message of messages) {
    if (message?.role !== 'assistant') {
      continue
    }
    const text = extractTextFromContent(message.content)
    if (text) {
      outputs.push(text)
    }
  }

  return outputs
}

export function extractPersistedMessageText(message: PersistedChatMessage): string {
  const raw = message.content
  if (!raw) return ''

  // If it's a string, it might be a JSON string or just text
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return extractTextFromContent(parsed?.content ?? parsed).trim()
    } catch {
      return raw.trim()
    }
  }

  if (typeof raw === 'object') {
    const obj = raw as Record<string, unknown>
    // The backend stores ModelMessage which has a 'content' field
    if ('content' in obj && obj.content !== undefined && obj.content !== null) {
      return extractTextFromContent(obj.content).trim()
    }
    return extractTextFromContent(raw).trim()
  }

  return extractTextFromContent(raw).trim()
}

function parseStreamPayload(payload: string): Record<string, unknown> | string | null {
  let current: unknown = payload
  for (let i = 0; i < 2; i += 1) {
    if (typeof current !== 'string') {
      break
    }
    const raw = current.trim()
    if (!raw || raw === '[DONE]') {
      return null
    }
    try {
      current = JSON.parse(raw)
      continue
    } catch {
      return raw
    }
  }

  if (typeof current === 'string') {
    return current.trim()
  }
  if (current && typeof current === 'object') {
    return current as Record<string, unknown>
  }
  return null
}

export function extractTextFromContent(content: unknown): string {
  if (typeof content === 'string') {
    return content.trim()
  }

  if (Array.isArray(content)) {
    const lines = content
      .map((part) => {
        if (!part || typeof part !== 'object') {
          return ''
        }

        const value = part as Record<string, unknown>
        const partType = String(value.type ?? '').toLowerCase()
        if (partType === 'text' && typeof value.text === 'string') {
          return value.text.trim()
        }
        if (partType === 'link' && typeof value.url === 'string') {
          return value.url.trim()
        }
        if (partType === 'emoji' && typeof value.emoji === 'string') {
          return value.emoji.trim()
        }
        if (typeof value.text === 'string') {
          return value.text.trim()
        }
        return ''
      })
      .filter(Boolean)

    return lines.join('\n').trim()
  }

  if (content && typeof content === 'object') {
    const value = content as Record<string, unknown>
    if (typeof value.text === 'string') {
      return value.text.trim()
    }
  }

  return ''
}

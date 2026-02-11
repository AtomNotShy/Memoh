import { tool } from 'ai'
import { z } from 'zod'
import { AuthFetcher } from '..'
import type { IdentityContext } from '../types'

export type MessageToolParams = {
  fetch: AuthFetcher
  identity: IdentityContext
}

const SendMessageSchema = z.object({
  bot_id: z.string().optional(),
  platform: z.string().optional(),
  target: z.string().optional(),
  channel_identity_id: z.string().optional(),
  to_user_id: z.string().optional(),
  message: z.string(),
})

export const getMessageTools = ({ fetch, identity }: MessageToolParams) => {
  const sendMessage = tool({
    description: 'Send a message to a channel or session',
    inputSchema: SendMessageSchema,
    execute: async (payload) => {
      const botId = (payload.bot_id ?? identity.botId ?? '').trim()
      const platform = (payload.platform ?? identity.currentPlatform ?? '').trim()
      const replyTarget = (identity.replyTarget ?? '').trim()
      const target = (payload.target ?? replyTarget).trim()
      const channelIdentityID = (payload.channel_identity_id ?? payload.to_user_id ?? '').trim()
      if (!botId) {
        throw new Error('bot_id is required')
      }
      if (!platform) {
        throw new Error('platform is required')
      }
      // Prefer chat token when there is no explicit target identity.
      const useSessionToken = !!identity.sessionToken && !channelIdentityID
      if (!target && !channelIdentityID && !useSessionToken) {
        throw new Error('target or channel_identity_id is required')
      }
      console.log('[Tool] send_message', {
        botId,
        platform,
        target: target || undefined,
        channelIdentityID: channelIdentityID || undefined,
        replyTarget,
        useSessionToken,
      })
      const body: Record<string, unknown> = {
        message: {
          text: payload.message,
        },
      }
      if (target) {
        body.target = target
      }
      if (channelIdentityID) {
        body.channel_identity_id = channelIdentityID
      }
      const url = useSessionToken
        ? `/bots/${botId}/channel/${platform}/send_chat`
        : `/bots/${botId}/channel/${platform}/send`
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (useSessionToken && identity.sessionToken) {
        headers.Authorization = `Bearer ${identity.sessionToken}`
      }
      const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      })
      const result = await response.json()
      return {
        ...result,
        instruction: 'Message delivered successfully. You have completed your response. Please STOP now and do not call any more tools.',
      }
    },
  })

  return {
    'send_message': sendMessage,
  }
}

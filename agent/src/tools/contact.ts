import { tool } from 'ai'
import { z } from 'zod'
import { AuthFetcher } from '..'
import type { IdentityContext } from '../types'

export type ContactToolParams = {
  fetch: AuthFetcher
  identity: IdentityContext
}

export const getContactTools = ({ fetch, identity }: ContactToolParams) => {
  const botId = identity.botId.trim()

  const listMyIdentities = async () => {
    const response = await fetch('/users/me/identities')
    return response.json()
  }

  const contactSearch = tool({
    description: 'Search identity cards by platform, external id, or display name',
    inputSchema: z.object({
      query: z.string().describe('The query to search identities').optional().default(''),
    }),
    execute: async ({ query }) => {
      const payload = await listMyIdentities()
      const keyword = query.trim().toLowerCase()
      const items = Array.isArray(payload?.items) ? payload.items : []
      const filtered = keyword
        ? items.filter((item: { platform?: string; external_id?: string; display_name?: string }) => {
          const platform = String(item?.platform ?? '').toLowerCase()
          const externalID = String(item?.external_id ?? '').toLowerCase()
          const displayName = String(item?.display_name ?? '').toLowerCase()
          return platform.includes(keyword) || externalID.includes(keyword) || displayName.includes(keyword)
        })
        : items
      return {
        canonical_channel_identity_id: payload?.canonical_channel_identity_id ?? '',
        total: filtered.length,
        items: filtered,
      }
    },
  })

  const contactCardMe = tool({
    description: 'Get my canonical identity card and all linked channel identities',
    inputSchema: z.object({}),
    execute: async () => {
      return listMyIdentities()
    },
  })

  const contactIssueBindCode = tool({
    description: 'Issue a bind code for linking current channel identity to this account',
    inputSchema: z.object({
      ttl_seconds: z.number().int().positive().optional().describe('Bind code ttl in seconds'),
    }),
    execute: async ({ ttl_seconds }) => {
      if (!botId) {
        throw new Error('bot_id is required')
      }
      const response = await fetch(`/bots/${botId}/bind_codes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ttl_seconds }),
      })
      return response.json()
    },
  })

  return {
    'contact_search': contactSearch,
    'contact_card_me': contactCardMe,
    'contact_issue_bind_code': contactIssueBindCode,
  }
}

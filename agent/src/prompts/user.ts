import { ContainerFileAttachment } from '../types'

export interface UserParams {
  channelIdentityId: string
  displayName: string
  channel: string
  date: Date
  attachments: ContainerFileAttachment[]
}

export const user = (
  query: string,
  { channelIdentityId, displayName, channel, date, attachments }: UserParams
) => {
  const headers = {
    'channel-identity-id': channelIdentityId,
    'display-name': displayName,
    'channel': channel,
    'time': date.toISOString(),
    'attachments': attachments.map(attachment => attachment.path),
  }
  return `
---
${Bun.YAML.stringify(headers)}
---
${query}
  `.trim()
}
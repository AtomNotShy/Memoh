import { ContainerFileAttachment } from '../types'

export interface UserParams {
  contactId: string
  contactName: string
  channel: string
  date: Date
  attachments: ContainerFileAttachment[]
}

export const user = (
  query: string,
  { contactId, contactName, channel, date, attachments }: UserParams
) => {
  const headers = {
    'contact-id': contactId,
    'contact-name': contactName,
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
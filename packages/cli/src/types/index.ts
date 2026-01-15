// API response type definitions

export interface ApiResponse<T = unknown> {
  success?: boolean
  data?: T
  error?: string
}

export interface User {
  id: string
  username: string
  role: string
  createdAt: string
}

export interface Model {
  id: string
  name: string
  modelId: string
  baseUrl: string
  apiKey?: string
  clientType: string
  type?: 'chat' | 'embedding'
  dimensions?: number
  createdAt: string
  updatedAt?: string
}

export interface Memory {
  content: string
  timestamp: string
  similarity?: number
}

export interface Message {
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export interface MessageListResponse {
  messages: Message[]
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
}

export interface Settings {
  userId: string
  language?: string
  maxContextLoadTime?: number
  defaultChatModel?: string
  defaultSummaryModel?: string
  defaultEmbeddingModel?: string
  createdAt: string
  updatedAt: string
}

export interface Schedule {
  id: string
  title: string
  description?: string
  cronExpression: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Platform {
  id: string
  name: string
  config: Record<string, unknown>
  active: boolean
  createdAt: string
  updatedAt: string
}

// Platform configuration definitions
export interface PlatformConfigField {
  name: string
  message: string
  type?: 'input' | 'password' | 'number'
  required?: boolean
  default?: string | number
  validate?: (value: string) => boolean | string
}

export interface PlatformDefinition {
  name: string
  displayName: string
  description: string
  configFields: PlatformConfigField[]
}

// Platform configurations
export const PLATFORM_DEFINITIONS: PlatformDefinition[] = [
  {
    name: 'telegram',
    displayName: 'Telegram',
    description: 'Telegram Bot Platform',
    configFields: [
      {
        name: 'botToken',
        message: 'Bot Token:',
        type: 'password',
        required: true,
        validate: (value: string) => {
          if (!value.trim()) return 'Bot token is required'
          return true
        },
      },
    ],
  },
  // Future platforms can be added here
  // {
  //   name: 'discord',
  //   displayName: 'Discord',
  //   description: 'Discord Bot Platform',
  //   configFields: [
  //     {
  //       name: 'botToken',
  //       message: 'Bot Token:',
  //       type: 'password',
  //       required: true,
  //     },
  //     {
  //       name: 'clientId',
  //       message: 'Client ID:',
  //       type: 'input',
  //       required: true,
  //     },
  //   ],
  // },
]

export interface MCPConnection {
  id: string
  type: string
  name: string
  config: MCPConnectionConfig
  active: boolean
  user: string
}

export type MCPConnectionConfig = 
  | StdioMCPConnection 
  | HTTPMCPConnection 
  | SSEMCPConnection

export interface StdioMCPConnection {
  type: 'stdio'
  command: string
  args: string[]
  env: Record<string, string>
  cwd: string
}

export interface HTTPMCPConnection {
  type: 'http'
  url: string
  headers: Record<string, string>
}

export interface SSEMCPConnection {
  type: 'sse'
  url: string
  headers: Record<string, string>
}


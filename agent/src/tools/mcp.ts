import { HTTPMCPConnection, MCPConnection, SSEMCPConnection, StdioMCPConnection } from '../types'
import { createMCPClient } from '@ai-sdk/mcp'

export const getMCPTools = async (connections: MCPConnection[]) => {
  const closeCallbacks: Array<() => Promise<void>> = []

  const getHTTPTools = async (connection: HTTPMCPConnection) => {
    const client = await createMCPClient({
      transport: {
        type: 'http',
        url: connection.url,
        headers: connection.headers,
      }
    })
    closeCallbacks.push(() => client.close())
    return await client.tools()
  }

  const getSSETools = async (connection: SSEMCPConnection) => {
    const client = await createMCPClient({
      transport: {
        type: 'sse',
        url: connection.url,
        headers: connection.headers,
      }
    })
    closeCallbacks.push(() => client.close())
    return await client.tools()
  }

  const getStdioTools = async (connection: StdioMCPConnection) => {
    // TODO: Implement stdio tools
    return {}
  }

  const toolSets = await Promise.all(connections.map(connection => {
    switch (connection.type) {
      case 'http':
        return getHTTPTools(connection)
      case 'sse':
        return getSSETools(connection)
      case 'stdio':
        return getStdioTools(connection)
    }
  }))

  return {
    tools: Object.assign({}, ...toolSets),
    close: async () => {
      await Promise.all(closeCallbacks.map(callback => callback()))
    }
  }
}
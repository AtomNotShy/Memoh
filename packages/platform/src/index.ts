export interface PlatformMessage {
  message: string
  userId: string
}

export class BasePlatform {
  name: string = 'base'
  description: string = 'Base platform'
  started: boolean = false

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async start(config: unknown): Promise<void> {}

  async stop(): Promise<void> {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async send(data: PlatformMessage): Promise<void> {}
}
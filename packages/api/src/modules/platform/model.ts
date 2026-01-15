import { z } from 'zod'

// Platform-specific config schemas
export const TelegramConfigSchema = z.object({
  botToken: z.string().min(1, 'Bot token is required'),
})

// Registry of platform config schemas
// When adding a new platform, add its config schema here
export const platformConfigSchemas: Record<string, z.ZodSchema> = {
  telegram: TelegramConfigSchema,
  // Add more platforms here as they are implemented
  // discord: DiscordConfigSchema,
  // slack: SlackConfigSchema,
}

// Helper function to get config schema for a platform
export const getPlatformConfigSchema = (platformName: string): z.ZodSchema => {
  const schema = platformConfigSchemas[platformName]
  if (!schema) {
    throw new Error(`Unknown platform: ${platformName}. Supported platforms: ${Object.keys(platformConfigSchemas).join(', ')}`)
  }
  return schema
}

// Base platform schema with dynamic config validation
const PlatformSchema = z.object({
  name: z.string().min(1, 'Platform name is required'),
  config: z.record(z.string(), z.unknown()),
  active: z.boolean().optional().default(true),
}).superRefine((data, ctx) => {
  // Validate that the platform name is supported
  if (!platformConfigSchemas[data.name]) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: `Unknown platform: ${data.name}. Supported platforms: ${Object.keys(platformConfigSchemas).join(', ')}`,
      path: ['name'],
    })
    return
  }

  // Validate the config against the platform-specific schema
  try {
    const configSchema = getPlatformConfigSchema(data.name)
    configSchema.parse(data.config)
  } catch (error) {
    if (error instanceof z.ZodError) {
      error.issues.forEach((issue: z.ZodIssue) => {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: issue.message,
          path: ['config', ...issue.path],
        })
      })
    }
  }
})

export type PlatformInput = z.infer<typeof PlatformSchema>

export const CreatePlatformModel = {
  body: PlatformSchema,
}

export const UpdatePlatformModel = {
  params: z.object({
    id: z.string(),
  }),
  body: PlatformSchema,
}

export const GetPlatformByIdModel = {
  params: z.object({
    id: z.string(),
  }),
}

export const DeletePlatformModel = {
  params: z.object({
    id: z.string(),
  }),
}

// For updating config, we need to know the platform name to validate
// This will be used with additional validation in the route handler
export const UpdatePlatformConfigModel = {
  params: z.object({
    id: z.string(),
  }),
  body: z.object({
    config: z.record(z.string(), z.unknown()),
  }),
}

export const SetPlatformActiveModel = {
  params: z.object({
    id: z.string(),
  }),
}


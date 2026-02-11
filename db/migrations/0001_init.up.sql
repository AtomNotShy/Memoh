CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('member', 'admin');
  END IF;
END
$$;

-- users: Memoh user principal
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT,
  email TEXT,
  password_hash TEXT,
  role user_role NOT NULL DEFAULT 'member',
  display_name TEXT,
  avatar_url TEXT,
  data_root TEXT,
  last_login_at TIMESTAMPTZ,
  chat_model_id TEXT,
  memory_model_id TEXT,
  embedding_model_id TEXT,
  max_context_load_time INTEGER NOT NULL DEFAULT 1440,
  language TEXT NOT NULL DEFAULT 'auto',
  is_active BOOLEAN NOT NULL DEFAULT true,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_email_unique UNIQUE (email),
  CONSTRAINT users_username_unique UNIQUE (username)
);

-- channel_identities: unified inbound identity subject
CREATE TABLE IF NOT EXISTS channel_identities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  channel TEXT NOT NULL,
  channel_subject_id TEXT NOT NULL,
  display_name TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT channel_identities_channel_subject_unique UNIQUE (channel, channel_subject_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_identities_user_id ON channel_identities(user_id);

-- user_channel_bindings: outbound delivery config
CREATE TABLE IF NOT EXISTS user_channel_bindings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  platform TEXT NOT NULL,
  config JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT user_channel_bindings_unique UNIQUE (user_id, platform)
);

CREATE INDEX IF NOT EXISTS idx_user_channel_bindings_user_id ON user_channel_bindings(user_id);

CREATE TABLE IF NOT EXISTS llm_providers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  client_type TEXT NOT NULL,
  base_url TEXT NOT NULL,
  api_key TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT llm_providers_name_unique UNIQUE (name),
  CONSTRAINT llm_providers_client_type_check CHECK (client_type IN ('openai', 'openai-compat', 'anthropic', 'google', 'ollama'))
);

CREATE TABLE IF NOT EXISTS models (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  model_id TEXT NOT NULL,
  name TEXT,
  llm_provider_id UUID NOT NULL REFERENCES llm_providers(id) ON DELETE CASCADE,
  dimensions INTEGER,
  is_multimodal BOOLEAN NOT NULL DEFAULT false,
  type TEXT NOT NULL DEFAULT 'chat',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT models_model_id_unique UNIQUE (model_id),
  CONSTRAINT models_type_check CHECK (type IN ('chat', 'embedding')),
  CONSTRAINT models_dimensions_check CHECK (type != 'embedding' OR dimensions IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS model_variants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  model_uuid UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
  variant_id TEXT NOT NULL,
  weight INTEGER NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_model_variants_model_uuid ON model_variants(model_uuid);
CREATE INDEX IF NOT EXISTS idx_model_variants_variant_id ON model_variants(variant_id);

CREATE TABLE IF NOT EXISTS bots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  display_name TEXT,
  avatar_url TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  max_context_load_time INTEGER NOT NULL DEFAULT 1440,
  language TEXT NOT NULL DEFAULT 'auto',
  allow_guest BOOLEAN NOT NULL DEFAULT false,
  chat_model_id UUID REFERENCES models(id) ON DELETE SET NULL,
  memory_model_id UUID REFERENCES models(id) ON DELETE SET NULL,
  embedding_model_id UUID REFERENCES models(id) ON DELETE SET NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT bots_type_check CHECK (type IN ('personal', 'public'))
);

CREATE INDEX IF NOT EXISTS idx_bots_owner_user_id ON bots(owner_user_id);

CREATE TABLE IF NOT EXISTS bot_members (
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'member',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT bot_members_role_check CHECK (role IN ('owner', 'admin', 'member')),
  CONSTRAINT bot_members_unique UNIQUE (bot_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_bot_members_user_id ON bot_members(user_id);

CREATE TABLE IF NOT EXISTS mcp_connections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  config JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT mcp_connections_type_check CHECK (type IN ('stdio', 'http', 'sse')),
  CONSTRAINT mcp_connections_unique UNIQUE (bot_id, name)
);

CREATE INDEX IF NOT EXISTS idx_mcp_connections_bot_id ON mcp_connections(bot_id);

-- chats: first-class conversation container
CREATE TABLE IF NOT EXISTS chats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('direct', 'group', 'thread')),
  parent_chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
  title TEXT,
  created_by_user_id UUID REFERENCES users(id),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  enable_chat_memory BOOLEAN NOT NULL DEFAULT true,
  enable_private_memory BOOLEAN NOT NULL DEFAULT true,
  enable_public_memory BOOLEAN NOT NULL DEFAULT false,
  model_id TEXT,
  settings_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chats_bot_id ON chats(bot_id);
CREATE INDEX IF NOT EXISTS idx_chats_parent ON chats(parent_chat_id);

-- chat_participants: chat membership
CREATE TABLE IF NOT EXISTS chat_participants (
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_participants_user ON chat_participants(user_id);

-- chat_messages: per-message storage (replaces history)
CREATE TABLE IF NOT EXISTS chat_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  route_id UUID,
  sender_channel_identity_id UUID REFERENCES channel_identities(id),
  sender_user_id UUID REFERENCES users(id),
  platform TEXT,
  external_message_id TEXT,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'tool')),
  content JSONB NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Backfill newly introduced columns for existing deployments where chat_messages
-- was created before route/platform traceability fields were added.
ALTER TABLE IF EXISTS chat_messages
  ADD COLUMN IF NOT EXISTS route_id UUID;

ALTER TABLE IF EXISTS chat_messages
  ADD COLUMN IF NOT EXISTS platform TEXT;

ALTER TABLE IF EXISTS chat_messages
  ADD COLUMN IF NOT EXISTS external_message_id TEXT;

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_created ON chat_messages(chat_id, created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_bot ON chat_messages(bot_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_route ON chat_messages(route_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_external_lookup
  ON chat_messages(platform, external_message_id);

-- chat_channel_identity_presence: derived cache of channel identities observed in chats
CREATE TABLE IF NOT EXISTS chat_channel_identity_presence (
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  channel_identity_id UUID NOT NULL REFERENCES channel_identities(id) ON DELETE CASCADE,
  first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  message_count BIGINT NOT NULL DEFAULT 1,
  PRIMARY KEY (chat_id, channel_identity_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_channel_identity_presence_identity_last_seen
  ON chat_channel_identity_presence(channel_identity_id, last_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_channel_identity_presence_chat_last_seen
  ON chat_channel_identity_presence(chat_id, last_seen_at DESC);

CREATE TABLE IF NOT EXISTS bot_channel_configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  channel_type TEXT NOT NULL,
  credentials JSONB NOT NULL DEFAULT '{}'::jsonb,
  external_identity TEXT,
  self_identity JSONB NOT NULL DEFAULT '{}'::jsonb,
  routing JSONB NOT NULL DEFAULT '{}'::jsonb,
  capabilities JSONB NOT NULL DEFAULT '{}'::jsonb,
  status TEXT NOT NULL DEFAULT 'pending',
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT bot_channel_status_check CHECK (status IN ('pending', 'verified', 'disabled')),
  CONSTRAINT bot_channel_unique UNIQUE (bot_id, channel_type)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_bot_channel_external_identity
  ON bot_channel_configs(channel_type, external_identity);

CREATE INDEX IF NOT EXISTS idx_bot_channel_bot_id ON bot_channel_configs(bot_id);

CREATE TABLE IF NOT EXISTS bot_preauth_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  token TEXT NOT NULL,
  issued_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  expires_at TIMESTAMPTZ,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT bot_preauth_keys_unique UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_bot_preauth_keys_bot_id ON bot_preauth_keys(bot_id);
CREATE INDEX IF NOT EXISTS idx_bot_preauth_keys_expires ON bot_preauth_keys(expires_at);

-- channel_identity_bind_codes: one-time codes for channel identity->user linking
CREATE TABLE IF NOT EXISTS channel_identity_bind_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  token TEXT NOT NULL,
  issued_by_user_id UUID NOT NULL REFERENCES users(id),
  platform TEXT,
  expires_at TIMESTAMPTZ,
  used_at TIMESTAMPTZ,
  used_by_channel_identity_id UUID REFERENCES channel_identities(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT channel_identity_bind_codes_token_unique UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_channel_identity_bind_codes_platform ON channel_identity_bind_codes(platform);

-- chat_routes: routing mapping (replaces channel_sessions)
CREATE TABLE IF NOT EXISTS chat_routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  platform TEXT NOT NULL,
  channel_config_id UUID REFERENCES bot_channel_configs(id) ON DELETE SET NULL,
  conversation_id TEXT NOT NULL,
  thread_id TEXT,
  reply_target TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_routes_unique
  ON chat_routes (bot_id, platform, conversation_id, COALESCE(thread_id, ''));
CREATE INDEX IF NOT EXISTS idx_chat_routes_chat ON chat_routes(chat_id);
CREATE INDEX IF NOT EXISTS idx_chat_routes_bot ON chat_routes(bot_id);

CREATE TABLE IF NOT EXISTS containers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  container_id TEXT NOT NULL,
  container_name TEXT NOT NULL,
  image TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'created',
  namespace TEXT NOT NULL DEFAULT 'default',
  auto_start BOOLEAN NOT NULL DEFAULT true,
  host_path TEXT,
  container_path TEXT NOT NULL DEFAULT '/data',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_started_at TIMESTAMPTZ,
  last_stopped_at TIMESTAMPTZ,
  CONSTRAINT containers_container_id_unique UNIQUE (container_id),
  CONSTRAINT containers_container_name_unique UNIQUE (container_name)
);

CREATE INDEX IF NOT EXISTS idx_containers_bot_id ON containers(bot_id);

CREATE TABLE IF NOT EXISTS snapshots (
  id TEXT PRIMARY KEY,
  container_id TEXT NOT NULL REFERENCES containers(container_id) ON DELETE CASCADE,
  parent_snapshot_id TEXT REFERENCES snapshots(id) ON DELETE SET NULL,
  snapshotter TEXT NOT NULL,
  digest TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_snapshots_container_id ON snapshots(container_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_parent_id ON snapshots(parent_snapshot_id);

CREATE TABLE IF NOT EXISTS container_versions (
  id TEXT PRIMARY KEY,
  container_id TEXT NOT NULL REFERENCES containers(container_id) ON DELETE CASCADE,
  snapshot_id TEXT NOT NULL REFERENCES snapshots(id) ON DELETE RESTRICT,
  version INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (container_id, version)
);

CREATE INDEX IF NOT EXISTS idx_container_versions_container_id ON container_versions(container_id);

CREATE TABLE IF NOT EXISTS lifecycle_events (
  id TEXT PRIMARY KEY,
  container_id TEXT NOT NULL REFERENCES containers(container_id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lifecycle_events_container_id ON lifecycle_events(container_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_events_event_type ON lifecycle_events(event_type);

CREATE TABLE IF NOT EXISTS schedule (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  pattern TEXT NOT NULL,
  max_calls INTEGER,
  current_calls INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  enabled BOOLEAN NOT NULL DEFAULT true,
  command TEXT NOT NULL,
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_schedule_bot_id ON schedule(bot_id);
CREATE INDEX IF NOT EXISTS idx_schedule_enabled ON schedule(enabled);

CREATE TABLE IF NOT EXISTS subagents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted BOOLEAN NOT NULL DEFAULT false,
  deleted_at TIMESTAMPTZ,
  bot_id UUID NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
  messages JSONB NOT NULL DEFAULT '[]'::jsonb,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  skills JSONB NOT NULL DEFAULT '[]'::jsonb,
  CONSTRAINT subagents_name_unique UNIQUE (bot_id, name)
);

CREATE INDEX IF NOT EXISTS idx_subagents_bot_id ON subagents(bot_id);
CREATE INDEX IF NOT EXISTS idx_subagents_deleted ON subagents(deleted);


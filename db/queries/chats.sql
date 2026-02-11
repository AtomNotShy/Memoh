-- name: CreateChat :one
INSERT INTO chats (bot_id, kind, parent_chat_id, title, created_by_user_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, bot_id, kind, parent_chat_id, title, created_by_user_id, metadata, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata, created_at, updated_at;

-- name: GetChatByID :one
SELECT id, bot_id, kind, parent_chat_id, title, created_by_user_id, metadata, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata, created_at, updated_at
FROM chats
WHERE id = $1;

-- name: ListChatsByBotAndUser :many
SELECT c.id, c.bot_id, c.kind, c.parent_chat_id, c.title, c.created_by_user_id, c.metadata, c.enable_chat_memory, c.enable_private_memory, c.enable_public_memory, c.model_id, c.settings_metadata, c.created_at, c.updated_at
FROM chats c
JOIN chat_participants cp ON cp.chat_id = c.id
WHERE c.bot_id = $1 AND cp.user_id = $2
ORDER BY c.updated_at DESC;

-- name: ListVisibleChatsByBotAndUser :many
WITH participant_chats AS (
  SELECT c.id, c.bot_id, c.kind, c.parent_chat_id, c.title, c.created_by_user_id, c.metadata, c.created_at, c.updated_at,
         'participant'::text AS access_mode,
         cp.role AS participant_role,
         NULL::timestamptz AS last_observed_at
  FROM chats c
  JOIN chat_participants cp ON cp.chat_id = c.id
  WHERE c.bot_id = $1 AND cp.user_id = $2
),
observed_chats AS (
  SELECT c.id, c.bot_id, c.kind, c.parent_chat_id, c.title, c.created_by_user_id, c.metadata, c.created_at, c.updated_at,
         'channel_identity_observed'::text AS access_mode,
         ''::text AS participant_role,
         MAX(cap.last_seen_at) AS last_observed_at
  FROM chats c
  JOIN chat_channel_identity_presence cap ON cap.chat_id = c.id
  JOIN channel_identities ci ON ci.id = cap.channel_identity_id
  WHERE c.bot_id = $1
    AND ci.user_id = $2
    AND NOT EXISTS (
      SELECT 1 FROM chat_participants cp
      WHERE cp.chat_id = c.id AND cp.user_id = $2
    )
  GROUP BY c.id, c.bot_id, c.kind, c.parent_chat_id, c.title, c.created_by_user_id, c.metadata, c.created_at, c.updated_at
)
SELECT id, bot_id, kind, parent_chat_id, title, created_by_user_id, metadata, created_at, updated_at,
       access_mode, participant_role, last_observed_at
FROM (
  SELECT * FROM participant_chats
  UNION ALL
  SELECT * FROM observed_chats
) v
ORDER BY v.updated_at DESC, v.last_observed_at DESC NULLS LAST;

-- name: GetChatReadAccessByUser :one
WITH participant_access AS (
  SELECT 'participant'::text AS access_mode,
         cp.role AS participant_role,
         NULL::timestamptz AS last_observed_at
  FROM chat_participants cp
  WHERE cp.chat_id = $1 AND cp.user_id = $2
),
observed_access AS (
  SELECT 'channel_identity_observed'::text AS access_mode,
         ''::text AS participant_role,
         MAX(cap.last_seen_at) AS last_observed_at
  FROM chat_channel_identity_presence cap
  JOIN channel_identities ci ON ci.id = cap.channel_identity_id
  WHERE cap.chat_id = $1 AND ci.user_id = $2
  GROUP BY cap.chat_id
),
all_access AS (
  SELECT * FROM participant_access
  UNION ALL
  SELECT * FROM observed_access
)
SELECT access_mode, participant_role, last_observed_at
FROM all_access
ORDER BY CASE WHEN access_mode = 'participant' THEN 0 ELSE 1 END, last_observed_at DESC NULLS LAST
LIMIT 1;

-- name: ListThreadsByParent :many
SELECT id, bot_id, kind, parent_chat_id, title, created_by_user_id, metadata, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata, created_at, updated_at
FROM chats
WHERE parent_chat_id = $1 AND kind = 'thread'
ORDER BY created_at DESC;

-- name: UpdateChatTitle :one
UPDATE chats SET title = $2, updated_at = now()
WHERE id = $1
RETURNING id, bot_id, kind, parent_chat_id, title, created_by_user_id, metadata, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata, created_at, updated_at;

-- name: TouchChat :exec
UPDATE chats SET updated_at = now() WHERE id = $1;

-- name: DeleteChat :exec
DELETE FROM chats WHERE id = $1;

-- chat_participants

-- name: AddChatParticipant :one
INSERT INTO chat_participants (chat_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (chat_id, user_id) DO UPDATE SET role = EXCLUDED.role
RETURNING chat_id, user_id, role, joined_at;

-- name: GetChatParticipant :one
SELECT chat_id, user_id, role, joined_at
FROM chat_participants
WHERE chat_id = $1 AND user_id = $2;

-- name: ListChatParticipants :many
SELECT chat_id, user_id, role, joined_at
FROM chat_participants
WHERE chat_id = $1
ORDER BY joined_at ASC;

-- name: RemoveChatParticipant :exec
DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2;

-- name: CopyParticipantsToChat :exec
INSERT INTO chat_participants (chat_id, user_id, role)
SELECT $2, cp.user_id, cp.role FROM chat_participants cp WHERE cp.chat_id = $1
ON CONFLICT (chat_id, user_id) DO NOTHING;

-- chat_settings

-- name: UpsertChatSettings :one
UPDATE chats
SET enable_chat_memory = $2,
    enable_private_memory = $3,
    enable_public_memory = $4,
    model_id = $5,
    settings_metadata = $6
WHERE id = $1
RETURNING id AS chat_id, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata AS metadata, updated_at;

-- name: GetChatSettings :one
SELECT id AS chat_id, enable_chat_memory, enable_private_memory, enable_public_memory, model_id, settings_metadata AS metadata, updated_at
FROM chats
WHERE id = $1;

-- chat_routes

-- name: CreateChatRoute :one
INSERT INTO chat_routes (chat_id, bot_id, platform, channel_config_id, conversation_id, thread_id, reply_target, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, chat_id, bot_id, platform, channel_config_id, conversation_id, thread_id, reply_target, metadata, created_at, updated_at;

-- name: FindChatRoute :one
SELECT id, chat_id, bot_id, platform, channel_config_id, conversation_id, thread_id, reply_target, metadata, created_at, updated_at
FROM chat_routes
WHERE bot_id = $1 AND platform = $2 AND conversation_id = $3
  AND COALESCE(thread_id, '') = COALESCE(sqlc.narg('thread_id'), '')
LIMIT 1;

-- name: GetChatRouteByID :one
SELECT id, chat_id, bot_id, platform, channel_config_id, conversation_id, thread_id, reply_target, metadata, created_at, updated_at
FROM chat_routes
WHERE id = $1;

-- name: ListChatRoutes :many
SELECT id, chat_id, bot_id, platform, channel_config_id, conversation_id, thread_id, reply_target, metadata, created_at, updated_at
FROM chat_routes
WHERE chat_id = $1
ORDER BY created_at ASC;

-- name: UpdateChatRouteReplyTarget :exec
UPDATE chat_routes SET reply_target = $2, updated_at = now() WHERE id = $1;

-- name: DeleteChatRoute :exec
DELETE FROM chat_routes WHERE id = $1;

-- chat_messages

-- name: CreateChatMessage :one
INSERT INTO chat_messages (chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata, created_at;

-- name: UpsertChatChannelIdentityPresence :exec
INSERT INTO chat_channel_identity_presence (chat_id, channel_identity_id, first_seen_at, last_seen_at, message_count)
VALUES ($1, $2, now(), now(), 1)
ON CONFLICT (chat_id, channel_identity_id)
DO UPDATE SET
  last_seen_at = now(),
  message_count = chat_channel_identity_presence.message_count + 1;

-- name: ListChatMessages :many
SELECT id, chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata, created_at
FROM chat_messages
WHERE chat_id = $1
ORDER BY created_at ASC;

-- name: ListChatMessagesSince :many
SELECT id, chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata, created_at
FROM chat_messages
WHERE chat_id = $1 AND created_at >= $2
ORDER BY created_at ASC;

-- name: ListChatMessagesBefore :many
SELECT id, chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata, created_at
FROM chat_messages
WHERE chat_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3;

-- name: ListChatMessagesLatest :many
SELECT id, chat_id, bot_id, route_id, sender_channel_identity_id, sender_user_id, platform, external_message_id, role, content, metadata, created_at
FROM chat_messages
WHERE chat_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: DeleteChatMessagesByChat :exec
DELETE FROM chat_messages WHERE chat_id = $1;

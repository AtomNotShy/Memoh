-- name: CreateChannelIdentity :one
INSERT INTO channel_identities (user_id, channel, channel_subject_id, display_name, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at;

-- name: GetChannelIdentityByID :one
SELECT id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at
FROM channel_identities
WHERE id = $1;

-- name: GetChannelIdentityByIDForUpdate :one
SELECT id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at
FROM channel_identities
WHERE id = $1
FOR UPDATE;

-- name: GetChannelIdentityByChannelSubject :one
SELECT id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at
FROM channel_identities
WHERE channel = $1 AND channel_subject_id = $2;

-- name: UpsertChannelIdentityByChannelSubject :one
INSERT INTO channel_identities (user_id, channel, channel_subject_id, display_name, metadata)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (channel, channel_subject_id)
DO UPDATE SET
  display_name = EXCLUDED.display_name,
  metadata = EXCLUDED.metadata,
  user_id = COALESCE(channel_identities.user_id, EXCLUDED.user_id),
  updated_at = now()
RETURNING id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at;

-- name: ListChannelIdentitiesByUserID :many
SELECT id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at
FROM channel_identities
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: SetChannelIdentityLinkedUser :one
UPDATE channel_identities
SET user_id = $2, updated_at = now()
WHERE id = $1
RETURNING id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at;

-- name: ClearChannelIdentityLinkedUser :one
UPDATE channel_identities
SET user_id = NULL, updated_at = now()
WHERE id = $1
RETURNING id, user_id, channel, channel_subject_id, display_name, metadata, created_at, updated_at;

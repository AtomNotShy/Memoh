-- name: GetSettingsByUserID :one
SELECT id AS user_id, chat_model_id, memory_model_id, embedding_model_id, max_context_load_time, language
FROM users
WHERE id = $1;

-- name: UpsertUserSettings :one
UPDATE users
SET chat_model_id = $2,
    memory_model_id = $3,
    embedding_model_id = $4,
    max_context_load_time = $5,
    language = $6,
    updated_at = now()
WHERE id = $1
RETURNING id AS user_id, chat_model_id, memory_model_id, embedding_model_id, max_context_load_time, language;

-- name: GetSettingsByBotID :one
SELECT
  bots.id AS bot_id,
  bots.max_context_load_time,
  bots.language,
  bots.allow_guest,
  chat_models.model_id AS chat_model_id,
  memory_models.model_id AS memory_model_id,
  embedding_models.model_id AS embedding_model_id
FROM bots
LEFT JOIN models AS chat_models ON chat_models.id = bots.chat_model_id
LEFT JOIN models AS memory_models ON memory_models.id = bots.memory_model_id
LEFT JOIN models AS embedding_models ON embedding_models.id = bots.embedding_model_id
WHERE bots.id = $1;

-- name: UpsertBotSettings :one
WITH updated AS (
  UPDATE bots
  SET max_context_load_time = sqlc.arg(max_context_load_time),
      language = sqlc.arg(language),
      allow_guest = sqlc.arg(allow_guest),
      chat_model_id = COALESCE(sqlc.narg(chat_model_id)::uuid, bots.chat_model_id),
      memory_model_id = COALESCE(sqlc.narg(memory_model_id)::uuid, bots.memory_model_id),
      embedding_model_id = COALESCE(sqlc.narg(embedding_model_id)::uuid, bots.embedding_model_id),
      updated_at = now()
  WHERE bots.id = sqlc.arg(id)
  RETURNING bots.id, bots.max_context_load_time, bots.language, bots.allow_guest, bots.chat_model_id, bots.memory_model_id, bots.embedding_model_id
)
SELECT
  updated.id AS bot_id,
  updated.max_context_load_time,
  updated.language,
  updated.allow_guest,
  chat_models.model_id AS chat_model_id,
  memory_models.model_id AS memory_model_id,
  embedding_models.model_id AS embedding_model_id
FROM updated
LEFT JOIN models AS chat_models ON chat_models.id = updated.chat_model_id
LEFT JOIN models AS memory_models ON memory_models.id = updated.memory_model_id
LEFT JOIN models AS embedding_models ON embedding_models.id = updated.embedding_model_id;

-- name: DeleteSettingsByBotID :exec
UPDATE bots
SET max_context_load_time = 1440,
    language = 'auto',
    allow_guest = false,
    updated_at = now()
WHERE id = $1;

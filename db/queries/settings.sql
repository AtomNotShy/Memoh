-- name: GetSettingsByUserID :one
SELECT user_id, chat_model_id, memory_model_id, embedding_model_id, max_context_load_time, language
FROM user_settings
WHERE user_id = $1;

-- name: UpsertUserSettings :one
INSERT INTO user_settings (user_id, chat_model_id, memory_model_id, embedding_model_id, max_context_load_time, language)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id) DO UPDATE SET
  chat_model_id = EXCLUDED.chat_model_id,
  memory_model_id = EXCLUDED.memory_model_id,
  embedding_model_id = EXCLUDED.embedding_model_id,
  max_context_load_time = EXCLUDED.max_context_load_time,
  language = EXCLUDED.language
RETURNING user_id, chat_model_id, memory_model_id, embedding_model_id, max_context_load_time, language;

-- name: GetSettingsByBotID :one
SELECT bot_id, max_context_load_time, language, allow_guest
FROM bot_settings
WHERE bot_id = $1;

-- name: GetBotModelConfigByBotID :one
SELECT
  bot_model_configs.bot_id,
  chat_models.model_id AS chat_model_id,
  memory_models.model_id AS memory_model_id,
  embedding_models.model_id AS embedding_model_id
FROM bot_model_configs
LEFT JOIN models AS chat_models ON chat_models.id = bot_model_configs.chat_model_id
LEFT JOIN models AS memory_models ON memory_models.id = bot_model_configs.memory_model_id
LEFT JOIN models AS embedding_models ON embedding_models.id = bot_model_configs.embedding_model_id
WHERE bot_model_configs.bot_id = $1;

-- name: UpsertBotSettings :one
INSERT INTO bot_settings (bot_id, max_context_load_time, language, allow_guest)
VALUES ($1, $2, $3, $4)
ON CONFLICT (bot_id) DO UPDATE SET
  max_context_load_time = EXCLUDED.max_context_load_time,
  language = EXCLUDED.language,
  allow_guest = EXCLUDED.allow_guest
RETURNING bot_id, max_context_load_time, language, allow_guest;

-- name: UpsertBotModelConfig :one
INSERT INTO bot_model_configs (bot_id, chat_model_id, memory_model_id, embedding_model_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (bot_id) DO UPDATE SET
  chat_model_id = COALESCE(EXCLUDED.chat_model_id, bot_model_configs.chat_model_id),
  memory_model_id = COALESCE(EXCLUDED.memory_model_id, bot_model_configs.memory_model_id),
  embedding_model_id = COALESCE(EXCLUDED.embedding_model_id, bot_model_configs.embedding_model_id)
RETURNING bot_id, chat_model_id, memory_model_id, embedding_model_id;

-- name: DeleteSettingsByBotID :exec
DELETE FROM bot_settings
WHERE bot_id = $1;


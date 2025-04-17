-- name: GetRecentMessages :many
SELECT id, content, created_at
FROM messages
ORDER BY created_at DESC
LIMIT 50;

-- name: CreateMessage :exec
INSERT INTO messages (content)
VALUES ($1);
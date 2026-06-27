-- name: CreateSession :one
INSERT INTO sessions (
    id, user_id, user_email, user_name, claims, refresh_token, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1;

-- name: TouchSession :one
UPDATE sessions
SET last_seen_at = now()
WHERE id = $1
RETURNING *;

-- name: RotateSession :one
-- Rotate the session id and refresh token (e.g. after a token refresh), and
-- slide the expiry forward.
UPDATE sessions
SET id = @new_id,
    refresh_token = @refresh_token,
    expires_at = @expires_at,
    last_seen_at = now()
WHERE id = @old_id
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions
WHERE expires_at < now();

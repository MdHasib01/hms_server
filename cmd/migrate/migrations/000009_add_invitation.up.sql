CREATE TABLE IF NOT EXISTS user_invitations (
  token bytea PRIMARY KEY,
  user_id uuid NOT NULL
)


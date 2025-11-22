-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    team_name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    FOREIGN KEY (team_name) REFERENCES teams(name) ON DELETE CASCADE
);

CREATE INDEX idx_users_user_id ON users (id);
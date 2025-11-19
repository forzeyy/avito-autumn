-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    team_name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    FOREIGN KEY (team_name) REFERENCES teams(name) ON DELETE CASCADE
);
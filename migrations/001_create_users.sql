-- Migration 001: Create users table
-- Pushpaka v1.0.0

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    api_key       UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    role          VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email   ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

COMMENT ON TABLE users IS 'Platform user accounts';

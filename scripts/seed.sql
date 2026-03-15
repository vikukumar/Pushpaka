-- Seed data for Pushpaka v1.0.0 demo
-- Run: psql $DATABASE_URL -f scripts/seed.sql

-- Demo user (password: Demo@1234)
-- bcrypt hash of 'Demo@1234'
INSERT INTO users (id, email, name, password_hash, api_key, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'demo@pushpaka.app',
    'Demo User',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    '00000000-0000-0000-0000-000000000099',
    'user'
) ON CONFLICT (email) DO NOTHING;

-- Demo project 1: Next.js app
INSERT INTO projects (id, user_id, name, repo_url, branch, build_command, start_command, port, framework, status)
VALUES (
    '00000000-0000-0000-0001-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'my-nextjs-app',
    'https://github.com/vercel/next.js',
    'main',
    'npm run build',
    'npm start',
    3000,
    'nextjs',
    'active'
) ON CONFLICT DO NOTHING;

-- Demo project 2: Go API
INSERT INTO projects (id, user_id, name, repo_url, branch, build_command, start_command, port, framework, status)
VALUES (
    '00000000-0000-0000-0001-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'go-api-service',
    'https://github.com/gin-gonic/gin',
    'master',
    'go build -o app .',
    './app',
    8080,
    'go',
    'active'
) ON CONFLICT DO NOTHING;

-- Demo deployment
INSERT INTO deployments (id, project_id, user_id, commit_sha, commit_msg, branch, status, image_tag, url, started_at, finished_at)
VALUES (
    '00000000-0000-0000-0002-000000000001',
    '00000000-0000-0000-0001-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'abc1234def5678',
    'feat: initial deployment',
    'main',
    'running',
    'pushpaka/00000000:latest',
    'https://my-nextjs-app.pushpaka.app',
    NOW() - INTERVAL '30 minutes',
    NOW() - INTERVAL '25 minutes'
) ON CONFLICT DO NOTHING;

-- Demo env vars
INSERT INTO environment_variables (project_id, user_id, key, value)
VALUES
    ('00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000001', 'NODE_ENV', 'production'),
    ('00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000001', 'NEXT_PUBLIC_API_URL', 'https://api.example.com')
ON CONFLICT (project_id, key) DO NOTHING;

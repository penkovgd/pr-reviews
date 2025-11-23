CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
CREATE TABLE teams (name VARCHAR(255) PRIMARY KEY);
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE TABLE pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(id),
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE
);
CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, user_id)
);
CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);
CREATE INDEX idx_pull_requests_created_at ON pull_requests(created_at);
CREATE INDEX idx_pull_request_reviewers_user_id ON pull_request_reviewers(user_id);
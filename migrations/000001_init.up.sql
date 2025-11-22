CREATE TABLE teams (
    name VARCHAR PRIMARY KEY
);

CREATE TABLE users (
    id VARCHAR PRIMARY KEY,
    username VARCHAR NOT NULL,
    team_name VARCHAR REFERENCES teams(name),
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE pull_requests (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    author_id VARCHAR REFERENCES users(id),
    status VARCHAR NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP
);

CREATE TABLE pr_reviewers (
    pr_id VARCHAR REFERENCES pull_requests(id),
    user_id VARCHAR REFERENCES users(id),
    PRIMARY KEY (pr_id, user_id)
);
-- +migrate Up
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id TEXT NOT NULL,
    reviewer_id TEXT NOT NULL,
    PRIMARY KEY (pr_id, reviewer_id),
    FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_id ON pr_reviewers (pr_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id_count ON pr_reviewers (reviewer_id) WHERE reviewer_id IS NOT NULL;
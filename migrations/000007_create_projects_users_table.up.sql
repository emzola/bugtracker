CREATE TABLE IF NOT EXISTS projects_users (
    project_id bigint NOT NULL REFERENCES projects ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    assigned_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);
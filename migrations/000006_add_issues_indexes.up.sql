CREATE INDEX IF NOT EXISTS issues_name_idx ON issues USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS issues_reported_date_idx ON issues USING GIN (reported_date);
CREATE INDEX IF NOT EXISTS issues_status_idx ON issues USING GIN (status);
CREATE INDEX IF NOT EXISTS issues_priority_idx ON issues USING GIN (priority);

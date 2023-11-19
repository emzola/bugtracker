CREATE INDEX IF NOT EXISTS projects_name_idx ON projects USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS projects_start_date_idx ON projects USING GIN (start_date);
CREATE INDEX IF NOT EXISTS projects_target_end_date_idx ON projects USING GIN (target_end_date);
CREATE INDEX IF NOT EXISTS projects_actual_end_date_idx ON projects USING GIN (actual_end_date);
CREATE INDEX IF NOT EXISTS projects_createdby_idx ON projects USING GIN (created_by);

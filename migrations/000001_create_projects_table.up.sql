CREATE TABLE IF NOT EXISTS projects (
    project_id bigserial PRIMARY KEY,
    project_name text UNIQUE NOT NULL,
    project_desc text NOT NULL DEFAULT '',
    start_date date NOT NULL,
    target_end_date date NOT NULL,
    actual_end_date date,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by text NOT NULL,
    modified_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    modified_by text NOT NULL,
    version integer NOT NULL DEFAULT 1
);
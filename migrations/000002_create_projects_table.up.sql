CREATE TABLE IF NOT EXISTS projects (
    id bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    description text NOT NULL DEFAULT '',
    assigned_to bigint REFERENCES users,
    start_date date NOT NULL,
    target_end_date date NOT NULL,
    actual_end_date date,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by text NOT NULL,
    modified_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    modified_by text NOT NULL,
    version integer NOT NULL DEFAULT 1
);
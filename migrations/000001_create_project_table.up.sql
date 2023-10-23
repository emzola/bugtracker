CREATE TABLE IF NOT EXISTS project (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    owner text NOT NULL,
    status text NOT NULL,
    start_date date,
    end_date date,
    completed_on date,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by text NOT NULL,
    modified_by text NOT NULL,
    public_access boolean NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1
);
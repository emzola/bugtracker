CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    role text NOT NULL,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by text NOT NULL,
    modified_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    modified_by text NOT NULL,
    version integer NOT NULL DEFAULT 1
);
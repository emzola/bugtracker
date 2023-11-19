CREATE EXTENSION citext;

CREATE TABLE IF NOT EXISTS people (
    person_id bigserial PRIMARY KEY,
    person_name text NOT NULL,
    person_email citext UNIQUE NOT NULL,
    person_role text NOT NULL,
    username text UNIQUE NOT NULL,
    created_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by text NOT NULL,
    modified_on timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    modified_by text NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);
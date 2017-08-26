CREATE TABLE users (
    id serial PRIMARY KEY,
    login text NOT NULL UNIQUE,
    password text NOT NULL
);

CREATE TABLE access_keys (
    user_id integer REFERENCES users ON DELETE CASCADE,
    key text NOT NULL,
    date_start date not null default CURRENT_DATE
);
CREATE INDEX index_access_keys ON access_keys (key);

CREATE TABLE history (
    id serial PRIMARY KEY,
    user_id integer REFERENCES users ON DELETE CASCADE,
    message text NOT NULL
);

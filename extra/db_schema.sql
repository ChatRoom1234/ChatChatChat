CREATE TABLE users (
    id serial PRIMARY KEY,
    login text NOT NULL UNIQUE,
    password text NOT NULL
);

CREATE TABLE rooms (
    id serial PRIMARY KEY,
    title text NOT NULL,
    url text NOT NULL
);
CREATE INDEX index_room_urls ON rooms (url);

CREATE TABLE access_keys (
    user_id integer REFERENCES users ON DELETE CASCADE,
    key text NOT NULL,
    date_start date not null default CURRENT_DATE
);
CREATE INDEX index_access_keys ON access_keys (key);

CREATE TABLE history (
    id serial PRIMARY KEY,
    message text NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY NOT NULL
);

CREATE TABLE IF NOT EXISTS segments (
    name VARCHAR(255) PRIMARY KEY NOT NULL,
    amount FLOAT
);

CREATE TABLE IF NOT EXISTS user_segments (
    user_id INTEGER NOT NULL,
    segment_name VARCHAR(255) NOT NULL,
    expire TIMESTAMP DEFAULT NULL,
    CONSTRAINT user_segments_pkey PRIMARY KEY (user_id, segment_name),
    CONSTRAINT user_segments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT user_segments_segment_name_fkey FOREIGN KEY (segment_name) REFERENCES segments (name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_segments_log (
    user_id INTEGER NOT NULL,
    segment_name VARCHAR(255) NOT NULL,
    operation VARCHAR(20) NOT NULL,
    operation_time TIMESTAMP DEFAULT NOW()
);
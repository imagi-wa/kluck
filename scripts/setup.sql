DROP TABLE channel_members;
DROP TABLE group_members;
DROP TABLE direct_messages;
DROP TABLE messages;
DROP TABLE channels;
DROP TABLE sessions;
DROP TABLE groups;
DROP TABLE types;
DROP TABLE users;
DROP TABLE status;

CREATE TABLE status (
    status_id INTEGER PRIMARY KEY,
    status_name varchar(8) NOT NULL
);

CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    user_name VARCHAR(32) NOT NULL DEFAULT 'kluck',
    user_code VARCHAR(20) UNIQUE,
    email_address VARCHAR(255) UNIQUE,
    password_hash VARCHAR(60),
    user_img TEXT DEFAULT 'default.png',
    enabled_otp BOOLEAN NOT NULL DEFAULT FALSE,
    otp_secret VARCHAR(32) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE types (
    type_id INTEGER PRIMARY KEY,
    type_name VARCHAR(32) NOT NULL
);

CREATE TABLE groups (
    group_id VARCHAR(20) PRIMARY KEY UNIQUE,
    group_name VARCHAR(32) NOT NULL,
    name_snippet VARCHAR(16) NOT NULL,
    members_number INTEGER DEFAULT 0,
    channels_number INTEGER DEFAULT 0,
    user_id INTEGER NOT NULL REFERENCES users(user_id) DEFAULT 0,
    type_id INTEGER NOT NULL REFERENCES types(type_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    session_id VARCHAR(36) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    authenticated BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE channels (
    channel_id VARCHAR(26) PRIMARY KEY,
    channel_name TEXT,
    channel_topic TEXT,
    name_snippet VARCHAR(64),
    topic_snippet VARCHAR(256),
    messages_number INTEGER NOT NULL DEFAULT 0,
    members_number INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER REFERENCES users(user_id) DEFAULT 0,
    group_id VARCHAR(20) REFERENCES groups(group_id),
    type_id INTEGER NOT NULL REFERENCES types(type_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    message_id SERIAL PRIMARY KEY,
    sentence TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    channel_id VARCHAR(26) NOT NULL REFERENCES channels(channel_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE direct_messages (
    direct_message_id SERIAL PRIMARY KEY,
    sentence TEXT NOT NULL,
    sender_id INTEGER NOT NULL REFERENCES users(user_id),
    receiver_id INTEGER NOT NULL REFERENCES users(user_id),
    group_id VARCHAR(20) NOT NULL REFERENCES groups(group_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE channel_members (
    member_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    channel_id VARCHAR(26) NOT NULL REFERENCES channels(channel_id),
    status_id INTEGER NOT NULL REFERENCES status(status_id),
    last_accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_members (
    member_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    group_id VARCHAR(20) NOT NULL REFERENCES groups(group_id),
    status_id INTEGER NOT NULL REFERENCES status(status_id),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users VALUES (0);
INSERT INTO types VALUES (1, 'public'), (2, 'private'), (3, 'secret');
INSERT INTO status VALUES (1, 'online'), (2, 'offline'), (3, 'joined'), (4, 'pending');
INSERT INTO groups VALUES ('kluck-public', 'Kluck-Public', 'Kluck-Public', 0, 0, 0, 1);

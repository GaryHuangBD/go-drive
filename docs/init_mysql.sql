-- DDL

CREATE TABLE drives
(
    name    VARCHAR(255),
    type    VARCHAR(32) NOT NULL,
    config  VARCHAR(4096) NOT NULL,
    enabled INTEGER,
    PRIMARY KEY (name)
);

CREATE TABLE path_mount
(
    `path`     VARCHAR(4096),
    name     VARCHAR(255),
    mount_at VARCHAR(4096) NOT NULL,
    PRIMARY KEY (name)
);

CREATE TABLE groups
(
    name VARCHAR(64),
     PRIMARY KEY(name)
);

CREATE TABLE path_permissions
(
  `path`     VARCHAR(4096),
    subject    VARCHAR(64) NOT NULL,
    permission INTEGER NOT NULL,
    policy     INTEGER NOT NULL,
    depth      INTEGER NOT NULL
);

CREATE TABLE user_groups
(
    group_name VARCHAR(64),
    username   VARCHAR(255),
    PRIMARY KEY (group_name, username)
);

CREATE TABLE users
(
    username VARCHAR(255),
    password VARCHAR(128) NOT NULL,
     PRIMARY KEY(username)
);

CREATE TABLE drive_data
(
    drive      VARCHAR(255),
    data_key   VARCHAR(255),
    data_value VARCHAR(4096),
    PRIMARY KEY (drive, data_key)
);

CREATE TABLE drive_cache
(
    drive       VARCHAR(255) NOT NULL,
    `path`        VARCHAR(255) NOT NULL,
   `depth`       INTEGER NOT NULL,
    type        INTEGER NOT NULL,
    cache_value TEXT    NOT NULL,
    expires_at  INTEGER NOT NULL,
    PRIMARY KEY (drive, `path`, `depth`, type)
);

-- Init data

INSERT INTO users(username, password)
VALUES ('admin', '$2y$10$Xqn8qV2D2KY2ceI5esM/JOiKTPKJFbkSzzuhce89BxygvCqnhyk3m');
-- 123456

INSERT INTO groups(name)
VALUES ('admin');

INSERT INTO user_groups(username, group_name)
VALUES ('admin', 'admin');

INSERT INTO path_permissions(path, subject, permission, policy, depth)
VALUES ('', 'ANY', 1, 1, 0);
INSERT INTO path_permissions(path, subject, permission, policy, depth)
VALUES ('', 'g:admin', 3, 1, 0);

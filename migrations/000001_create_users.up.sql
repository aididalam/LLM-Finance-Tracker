CREATE TABLE users (
    user_id           CHAR(36) NOT NULL,
    first_name        VARCHAR(128) NOT NULL DEFAULT '',
    last_name         VARCHAR(128) NOT NULL DEFAULT '',
    display_name      VARCHAR(255) NOT NULL DEFAULT '',
    email             VARCHAR(255) NOT NULL,
    password_hash     VARCHAR(255) NOT NULL DEFAULT '',
    status            ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
    avatar_url        VARCHAR(500) NOT NULL DEFAULT '',
    email_verified_at DATETIME NULL,
    last_login_at     DATETIME NULL,
    is_deleted        BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at        DATETIME NULL,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    UNIQUE KEY uq_users_email (email),
    INDEX idx_users_status (status, is_deleted)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO users (
    user_id, first_name, last_name, display_name, email, password_hash,
    status, avatar_url, email_verified_at
) VALUES (
    '00000000-0000-4000-8000-000000000001',
    'Demo',
    'User',
    'Demo User',
    'demo.user@example.com',
    '',
    'active',
    '',
    NOW()
);

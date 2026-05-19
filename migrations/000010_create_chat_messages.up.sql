CREATE TABLE chat_messages (
    id         BIGINT NOT NULL AUTO_INCREMENT,
    user_id    CHAR(36) NOT NULL,
    role       VARCHAR(20) NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_chat_messages_user_created (user_id, created_at),
    CONSTRAINT fk_chat_messages_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

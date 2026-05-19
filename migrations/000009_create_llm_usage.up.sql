CREATE TABLE llm_usage (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id       CHAR(36) NOT NULL,
    provider      VARCHAR(32) NOT NULL,
    model         VARCHAR(64) NOT NULL,
    prompt_tokens INT UNSIGNED NOT NULL DEFAULT 0,
    output_tokens INT UNSIGNED NOT NULL DEFAULT 0,
    purpose       VARCHAR(64) NOT NULL DEFAULT 'expense_parse',
    receipt_type  ENUM('image', 'pdf') NULL DEFAULT NULL,
    cost_usd      DECIMAL(10, 8) NOT NULL DEFAULT 0.00000000,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_created (created_at),
    INDEX idx_llm_usage_user_created (user_id, created_at),
    CONSTRAINT fk_llm_usage_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

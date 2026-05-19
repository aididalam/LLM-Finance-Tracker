CREATE TABLE users_settings (
    user_id                CHAR(36) NOT NULL,
    currency               CHAR(3) NOT NULL DEFAULT 'BDT',
    budget_alert_threshold TINYINT UNSIGNED NOT NULL DEFAULT 80,
    budget_alert_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    timezone               VARCHAR(64) NOT NULL DEFAULT 'Asia/Dhaka',
    locale                 VARCHAR(16) NOT NULL DEFAULT 'en-BD',
    telegram_chat_id       VARCHAR(64) NULL,
    metadata               JSON NULL,
    updated_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    UNIQUE KEY uq_users_settings_telegram_chat_id (telegram_chat_id),
    CONSTRAINT fk_users_settings_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO users_settings (
    user_id, currency, budget_alert_threshold, budget_alert_enabled,
    timezone, locale, metadata
) VALUES (
    '00000000-0000-4000-8000-000000000001',
    'BDT',
    80,
    TRUE,
    'Asia/Dhaka',
    'en-BD',
    JSON_OBJECT('bootstrap', true, 'login_enabled', false)
);

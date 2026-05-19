CREATE TABLE wallet (
    id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    name VARCHAR(150) NOT NULL,
    account_type ENUM('cash', 'bank', 'mfs') NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_wallet_user_id (user_id),
    INDEX idx_wallet_account_type (account_type),
    CONSTRAINT fk_wallet_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO
    wallet (id, user_id, name, account_type, is_active)
VALUES
    (
        '00000000-0000-4000-8000-000000000101',
        '00000000-0000-4000-8000-000000000001',
        'Cash Wallet',
        'cash',
        TRUE
    );
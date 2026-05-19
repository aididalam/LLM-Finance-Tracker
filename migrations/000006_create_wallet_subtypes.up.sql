CREATE TABLE wallet_mfs (
    id CHAR(36) NOT NULL,
    wallet_id CHAR(36) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    mfs_id CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_wallet_mfs_wallet_id (wallet_id),
    CONSTRAINT fk_wallet_mfs_wallet FOREIGN KEY (wallet_id) REFERENCES wallet(id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE wallet_bank (
    id CHAR(36) NOT NULL,
    wallet_id CHAR(36) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    bank_name VARCHAR(150) NOT NULL,
    branch VARCHAR(150) NULL,
    account_numeber VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_wallet_bank_wallet_id (wallet_id),
    UNIQUE KEY uq_wallet_bank_id_wallet_id (id, wallet_id),
    CONSTRAINT fk_wallet_bank_wallet FOREIGN KEY (wallet_id) REFERENCES wallet(id) ON DELETE CASCADE,
    CONSTRAINT chk_wallet_bank_account_numeber CHECK (CHAR_LENGTH(account_numeber) >= 4)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE wallet_bank_debit_card (
    id CHAR(36) NOT NULL,
    wallet_bank_id CHAR(36) NOT NULL,
    wallet_id CHAR(36) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_4_digit CHAR(4) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_wallet_bank_debit_card_wallet_bank_id (wallet_bank_id),
    INDEX idx_wallet_bank_debit_card_wallet_id (wallet_id),
    UNIQUE KEY uq_wallet_bank_debit_card_id_wallet_id (id, wallet_id),
    CONSTRAINT fk_wallet_bank_debit_card_bank FOREIGN KEY (wallet_bank_id, wallet_id) REFERENCES wallet_bank(id, wallet_id) ON DELETE CASCADE,
    CONSTRAINT chk_wallet_bank_debit_card_last_4_digit CHECK (last_4_digit REGEXP '^[0-9]{4}$')
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
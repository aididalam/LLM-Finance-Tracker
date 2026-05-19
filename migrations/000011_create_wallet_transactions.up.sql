CREATE TABLE wallet_expense_transctions (
    id CHAR(36) NOT NULL,
    wallet_id CHAR(36) NOT NULL,
    expense_id CHAR(36) NOT NULL,
    wallet_bank_debit_card_id CHAR(36) NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at DATETIME NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_wallet_expense_transctions_expense_id (expense_id),
    INDEX idx_wallet_expense_transctions_wallet_id (wallet_id),
    INDEX idx_wallet_expense_transctions_debit_card_wallet (wallet_bank_debit_card_id, wallet_id),
    CONSTRAINT fk_wallet_expense_transctions_wallet FOREIGN KEY (wallet_id) REFERENCES wallet(id),
    CONSTRAINT fk_wallet_expense_transctions_expense FOREIGN KEY (expense_id) REFERENCES expense(id) ON DELETE CASCADE,
    CONSTRAINT fk_wallet_expense_transctions_debit_card FOREIGN KEY (wallet_bank_debit_card_id, wallet_id) REFERENCES wallet_bank_debit_card(id, wallet_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE wallet_internal_transctions (
    id CHAR(36) NOT NULL,
    wallet_id CHAR(36) NOT NULL,
    source_type ENUM('initial', 'income', 'internal') NOT NULL,
    to_wallet_id CHAR(36) NULL,
    deleted_at DATETIME NULL,
    amount DECIMAL(15, 2) NOT NULL,
    fees DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_wallet_internal_transctions_wallet_id (wallet_id),
    INDEX idx_wallet_internal_transctions_to_wallet_id (to_wallet_id),
    INDEX idx_wallet_internal_transctions_source_type (source_type),
    CONSTRAINT fk_wallet_internal_transctions_wallet FOREIGN KEY (wallet_id) REFERENCES wallet(id),
    CONSTRAINT fk_wallet_internal_transctions_to_wallet FOREIGN KEY (to_wallet_id) REFERENCES wallet(id),
    CONSTRAINT chk_wallet_internal_transctions_amount CHECK (amount >= 0),
    CONSTRAINT chk_wallet_internal_transctions_fees CHECK (fees >= 0),
    CONSTRAINT chk_wallet_internal_transctions_to_wallet_required CHECK (
        (
            source_type = 'internal'
            AND to_wallet_id IS NOT NULL
        )
        OR (
            source_type IN ('initial', 'income')
            AND to_wallet_id IS NULL
        )
    ),
    CONSTRAINT chk_wallet_internal_transctions_not_same_wallet CHECK (
        to_wallet_id IS NULL
        OR wallet_id <> to_wallet_id
    )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
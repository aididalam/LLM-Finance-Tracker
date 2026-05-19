CREATE TABLE budgets (
    budget_id   CHAR(36) NOT NULL,
    user_id     CHAR(36) NOT NULL,
    category_id CHAR(36) NULL,
    amount      DECIMAL(15, 4) NOT NULL,
    month       TINYINT UNSIGNED NOT NULL,
    year        SMALLINT UNSIGNED NOT NULL,
    carry_over  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (budget_id),
    UNIQUE KEY uq_budget (category_id, month, year),
    UNIQUE KEY uq_budgets_user_category_month_year (user_id, category_id, month, year),
    INDEX idx_budgets_user (user_id),
    CONSTRAINT fk_budgets_category FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE,
    CONSTRAINT fk_budgets_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_budgets_user_category FOREIGN KEY (user_id, category_id) REFERENCES categories(user_id, category_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

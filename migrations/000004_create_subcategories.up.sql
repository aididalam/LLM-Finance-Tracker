CREATE TABLE subcategories (
    subcategory_id CHAR(36) NOT NULL,
    user_id        CHAR(36) NOT NULL,
    category_id    CHAR(36) NOT NULL,
    name           VARCHAR(128) NOT NULL,
    is_deleted     BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at     DATETIME NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (subcategory_id),
    UNIQUE KEY uq_subcategory_name (category_id, name),
    UNIQUE KEY uq_subcategories_user_subcategory (user_id, subcategory_id),
    INDEX idx_subcategory_category (category_id),
    INDEX idx_subcategories_user (user_id),
    CONSTRAINT fk_subcategories_category FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE,
    CONSTRAINT fk_subcategories_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_subcategories_user_category FOREIGN KEY (user_id, category_id) REFERENCES categories(user_id, category_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE categories (
    category_id CHAR(36) NOT NULL,
    user_id     CHAR(36) NOT NULL,
    name        VARCHAR(128) NOT NULL,
    icon        VARCHAR(32) NOT NULL DEFAULT '',
    color       CHAR(7) NOT NULL DEFAULT '',
    is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at  DATETIME NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (category_id),
    INDEX idx_categories_user (user_id),
    UNIQUE KEY uq_categories_user_name (user_id, name),
    UNIQUE KEY uq_categories_user_category (user_id, category_id),
    CONSTRAINT fk_categories_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

INSERT INTO categories (category_id, user_id, name, icon, color)
VALUES
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Food', '🍔', '#FF6B6B'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Transport', '🚗', '#4ECDC4'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Shopping', '🛍️', '#45B7D1'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Entertainment', '🎬', '#96CEB4'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Health', '💊', '#FFEAA7'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Bills', '📄', '#DDA0DD'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Education', '📚', '#98D8C8'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Salary', '💼', '#22C55E'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Freelance', '🧑‍💻', '#14B8A6'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Business', '🏢', '#0EA5E9'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Gift', '🎁', '#A855F7'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Interest', '🏦', '#84CC16'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Refund', '↩️', '#F59E0B'),
    (UUID(), '00000000-0000-4000-8000-000000000001', 'Other', '📦', '#B0B0B0');

CREATE TABLE hrd_deduction_categories (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 name VARCHAR(255) NOT NULL UNIQUE,
 description TEXT NOT NULL,
 sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 1,
 is_active BOOLEAN NOT NULL DEFAULT 1
);
CREATE TABLE hrd_deduction_rules (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 category_id BIGINT UNSIGNED NOT NULL,
 condition_text VARCHAR(1000) NOT NULL,
 deduction_percentage DECIMAL(5,2) NULL,
 is_sanction BOOLEAN NOT NULL DEFAULT 0,
 sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 1,
 is_active BOOLEAN NOT NULL DEFAULT 1,
 CONSTRAINT hrd_rule_category_fk FOREIGN KEY(category_id) REFERENCES hrd_deduction_categories(id),
 CONSTRAINT hrd_rule_percentage CHECK(deduction_percentage IS NULL OR deduction_percentage BETWEEN 0 AND 100),
 CONSTRAINT hrd_rule_kind CHECK((is_sanction=1 AND deduction_percentage IS NULL) OR (is_sanction=0 AND deduction_percentage IS NOT NULL))
);

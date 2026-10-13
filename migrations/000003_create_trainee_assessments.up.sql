-- Form penilaian apresiasi trainee; master rubrik dapat dipakai ulang.
CREATE TABLE trainee_assessment_aspects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    sort_order SMALLINT UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE trainee_assessment_criteria (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    aspect_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    weight DECIMAL(8,2) NOT NULL DEFAULT 1,
    sort_order SMALLINT UNSIGNED NOT NULL,
    CONSTRAINT trainee_criteria_weight_check CHECK (weight > 0),
    FOREIGN KEY (aspect_id) REFERENCES trainee_assessment_aspects(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE trainee_assessment_score_rules (
    criterion_id BIGINT UNSIGNED NOT NULL,
    score TINYINT UNSIGNED NOT NULL,
    finding_min SMALLINT UNSIGNED NOT NULL,
    finding_max SMALLINT UNSIGNED NULL,
    description VARCHAR(150) NOT NULL,
    PRIMARY KEY (criterion_id, score),
    CONSTRAINT trainee_rules_score_check CHECK (score BETWEEN 1 AND 5),
    CONSTRAINT trainee_rules_range_check CHECK (finding_max IS NULL OR finding_max >= finding_min),
    FOREIGN KEY (criterion_id) REFERENCES trainee_assessment_criteria(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE trainee_assessments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    trainee_user_id BIGINT UNSIGNED NOT NULL,
    assessor_user_id BIGINT UNSIGNED NOT NULL,
    store_id BIGINT UNSIGNED NOT NULL,
    employee_name VARCHAR(150) NOT NULL,
    employee_number VARCHAR(50) NOT NULL,
    department_name VARCHAR(150) NOT NULL,
    store_name VARCHAR(150) NOT NULL,
    training_month SMALLINT UNSIGNED NOT NULL,
    assessment_date DATE NOT NULL,
    notes TEXT NULL,
    status ENUM('draft','final') NOT NULL DEFAULT 'draft',
    signing_city VARCHAR(100) NOT NULL DEFAULT 'Yogyakarta',
    assessor_name VARCHAR(150) NOT NULL,
    assessor_signed_at DATETIME(6) NULL,
    employee_signed_at DATETIME(6) NULL,
    assessor_signature_path VARCHAR(255) NULL,
    employee_signature_path VARCHAR(255) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY trainee_assessment_month_unique (trainee_user_id, training_month),
    KEY trainee_assessment_store_date (store_id, assessment_date),
    CONSTRAINT trainee_assessment_month_check CHECK (training_month > 0),
    CONSTRAINT trainee_assessment_notes_check CHECK (status <> 'final' OR (notes IS NOT NULL AND CHAR_LENGTH(TRIM(notes)) > 0)),
    FOREIGN KEY (trainee_user_id) REFERENCES users(id),
    FOREIGN KEY (assessor_user_id) REFERENCES users(id),
    FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Snapshot teks/bobot menjaga riwayat ketika master rubrik berubah.
CREATE TABLE trainee_assessment_details (
    assessment_id BIGINT UNSIGNED NOT NULL,
    criterion_id BIGINT UNSIGNED NOT NULL,
    aspect_name VARCHAR(100) NOT NULL,
    criterion_description TEXT NOT NULL,
    weight DECIMAL(8,2) NOT NULL DEFAULT 1,
    score TINYINT UNSIGNED NULL,
    finding_count SMALLINT UNSIGNED NULL,
    evidence_notes TEXT NULL,
    PRIMARY KEY (assessment_id, criterion_id),
    CONSTRAINT trainee_details_score_check CHECK (score IS NULL OR score BETWEEN 1 AND 5),
    CONSTRAINT trainee_details_weight_check CHECK (weight > 0),
    FOREIGN KEY (assessment_id) REFERENCES trainee_assessments(id) ON DELETE CASCADE,
    FOREIGN KEY (criterion_id) REFERENCES trainee_assessment_criteria(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE VIEW trainee_assessment_results AS
SELECT totals.*,
    CASE WHEN average_score <= 1.80 THEN 'Sangat Kurang Sekali'
         WHEN average_score <= 2.60 THEN 'Kurang Sekali'
         WHEN average_score <= 3.40 THEN 'Kurang'
         WHEN average_score <= 4.20 THEN 'Mendekati Standar'
         WHEN average_score <= 5.00 THEN 'Sesuai Standar'
         ELSE NULL END AS rating
FROM (
    SELECT a.id AS assessment_id, COUNT(d.criterion_id) AS criterion_count,
        COUNT(d.score) AS scored_count, SUM(d.weight) AS total_weight,
        SUM(d.weight * d.score) AS total_score,
        CASE WHEN COUNT(d.criterion_id) > 0 AND COUNT(d.score) = COUNT(d.criterion_id)
             THEN ROUND(SUM(d.weight * d.score) / NULLIF(SUM(d.weight), 0), 2)
             ELSE NULL END AS average_score
    FROM trainee_assessments a
    LEFT JOIN trainee_assessment_details d ON d.assessment_id = a.id
    GROUP BY a.id
) totals;

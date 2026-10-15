-- Existing criteria and assessments retain NULL rubric_id: legacy rubric, not
-- retroactively assigned to a new rubric based only on training_month.
CREATE TABLE trainee_assessment_rubrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    department_id BIGINT UNSIGNED NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(150) NOT NULL,
    month_from SMALLINT UNSIGNED NOT NULL,
    month_to SMALLINT UNSIGNED NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE KEY rubric_department_code (department_id, code),
    KEY rubric_department_month (department_id, is_active, month_from, month_to),
    CHECK (month_from > 0),
    CHECK (month_to IS NULL OR month_to >= month_from),
    FOREIGN KEY (department_id) REFERENCES departments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE trainee_assessment_criteria
    ADD COLUMN rubric_id BIGINT UNSIGNED NULL,
    ADD CONSTRAINT fk_criteria_rubric FOREIGN KEY (rubric_id) REFERENCES trainee_assessment_rubrics(id),
    DROP INDEX code,
    ADD UNIQUE KEY criterion_rubric_code (rubric_id, code);

ALTER TABLE trainee_assessments
    ADD COLUMN department_id BIGINT UNSIGNED NULL,
    ADD COLUMN rubric_id BIGINT UNSIGNED NULL,
    ADD CONSTRAINT fk_assessment_department FOREIGN KEY (department_id) REFERENCES departments(id),
    ADD CONSTRAINT fk_assessment_rubric FOREIGN KEY (rubric_id) REFERENCES trainee_assessment_rubrics(id);

-- Use the saved assignment, never the employee's current department.
UPDATE trainee_assessments a
JOIN employee_assignments ea ON ea.id = a.trainee_assignment_id
SET a.department_id = ea.department_id, a.updated_at = a.updated_at
WHERE a.department_id IS NULL;

ALTER TABLE trainee_assessment_details
    ADD COLUMN criterion_code VARCHAR(50) NULL,
    ADD COLUMN score_rule_description VARCHAR(150) NULL,
    ADD COLUMN finding_min SMALLINT UNSIGNED NULL,
    ADD COLUMN finding_max SMALLINT UNSIGNED NULL;
-- Historical rule text cannot be reconstructed reliably; leave it NULL.

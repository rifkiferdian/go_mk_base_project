ALTER TABLE trainee_assessment_rubrics ADD COLUMN position_id BIGINT UNSIGNED NULL AFTER department_id;
ALTER TABLE trainee_assessment_criteria ADD COLUMN position_id BIGINT UNSIGNED NULL AFTER department_id;
ALTER TABLE trainee_assessments ADD COLUMN position_id BIGINT UNSIGNED NULL AFTER department_id;

UPDATE trainee_assessment_rubrics r
JOIN positions p ON p.department_id=r.department_id
SET r.position_id=p.id
WHERE p.code='customer_service' AND r.code LIKE 'cs_obat_%';
UPDATE trainee_assessment_criteria c
JOIN positions p ON p.department_id=c.department_id
SET c.position_id=p.id
WHERE p.code='customer_service' AND c.code LIKE 'cs_%';

UPDATE trainee_assessment_rubrics r
SET r.position_id=(SELECT MIN(p.id) FROM positions p WHERE p.department_id=r.department_id)
WHERE r.position_id IS NULL;
UPDATE trainee_assessment_criteria c
SET c.position_id=(SELECT MIN(p.id) FROM positions p WHERE p.department_id=c.department_id)
WHERE c.position_id IS NULL;
UPDATE trainee_assessments a
JOIN employee_assignments ea ON ea.id=a.trainee_assignment_id
SET a.position_id=ea.position_id
WHERE a.position_id IS NULL;

ALTER TABLE trainee_assessment_rubrics
    MODIFY position_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT fk_assessment_rubric_position FOREIGN KEY(position_id) REFERENCES positions(id),
    ADD KEY rubric_position_month(position_id,is_active,month_from,month_to);
ALTER TABLE trainee_assessment_criteria
    MODIFY position_id BIGINT UNSIGNED NOT NULL,
    ADD CONSTRAINT fk_assessment_criterion_position FOREIGN KEY(position_id) REFERENCES positions(id),
    ADD KEY criterion_position_sort(position_id,aspect_id,sort_order);
ALTER TABLE trainee_assessments
    ADD CONSTRAINT fk_trainee_assessment_position FOREIGN KEY(position_id) REFERENCES positions(id);

ALTER TABLE trainee_assessment_rubrics DROP INDEX rubric_department_code,
    ADD UNIQUE KEY rubric_position_code(position_id,code);
ALTER TABLE trainee_assessment_criteria DROP INDEX criterion_department_code,
    ADD UNIQUE KEY criterion_position_code(position_id,code);

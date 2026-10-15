-- Only unused schema can be rolled back. Seeded/used rubrics require a reviewed
-- data migration; the guard executes before any ALTER, even with mysql --force.
DELIMITER //
CREATE PROCEDURE rollback_department_rubrics()
BEGIN
    IF EXISTS (SELECT 1 FROM trainee_assessment_rubrics) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Rollback refused: preserve/export rubric data first';
    ELSE
        ALTER TABLE trainee_assessment_criteria ADD UNIQUE KEY code (code);
        ALTER TABLE trainee_assessment_details
            DROP COLUMN criterion_code, DROP COLUMN score_rule_description,
            DROP COLUMN finding_min, DROP COLUMN finding_max;
        ALTER TABLE trainee_assessments
            DROP FOREIGN KEY fk_assessment_department, DROP FOREIGN KEY fk_assessment_rubric,
            DROP COLUMN department_id, DROP COLUMN rubric_id;
        ALTER TABLE trainee_assessment_criteria
            DROP FOREIGN KEY fk_criteria_rubric,
            DROP INDEX criterion_rubric_code, DROP COLUMN rubric_id;
        DROP TABLE trainee_assessment_rubrics;
    END IF;
END//
DELIMITER ;
CALL rollback_department_rubrics();
DROP PROCEDURE rollback_department_rubrics;

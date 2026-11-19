ALTER TABLE trainee_assessment_criteria DROP INDEX criterion_position_code,
    ADD UNIQUE KEY criterion_department_code(department_id,code),
    DROP FOREIGN KEY fk_assessment_criterion_position,
    DROP INDEX criterion_position_sort,
    DROP COLUMN position_id;
ALTER TABLE trainee_assessment_rubrics DROP INDEX rubric_position_code,
    ADD UNIQUE KEY rubric_department_code(department_id,code),
    DROP FOREIGN KEY fk_assessment_rubric_position,
    DROP INDEX rubric_position_month,
    DROP COLUMN position_id;
ALTER TABLE trainee_assessments DROP FOREIGN KEY fk_trainee_assessment_position,DROP COLUMN position_id;

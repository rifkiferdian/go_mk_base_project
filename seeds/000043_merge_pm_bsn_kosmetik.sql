-- Gabungkan PM BSN KOSMETIK ke PM KOSMETIK.
START TRANSACTION;

SET @source_position_id := (
  SELECT id FROM positions WHERE code='pm_bsn_kosmetik' LIMIT 1
);
SET @target_position_id := (
  SELECT id FROM positions WHERE code='pm_kosmetik' LIMIT 1
);
SET @target_department_id := (
  SELECT department_id FROM positions WHERE id=@target_position_id LIMIT 1
);

UPDATE employee_assignments
SET position_id=@target_position_id,department_id=@target_department_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

UPDATE employee_placement_change_requests
SET requested_position_id=@target_position_id
WHERE requested_position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

UPDATE trainee_assessments a
JOIN trainee_assessment_rubrics source_rubric ON source_rubric.id=a.rubric_id
JOIN trainee_assessment_rubrics target_rubric
  ON target_rubric.position_id=@target_position_id
 AND target_rubric.month_from=source_rubric.month_from
 AND COALESCE(target_rubric.month_to,0)=COALESCE(source_rubric.month_to,0)
SET a.position_id=@target_position_id,a.rubric_id=target_rubric.id
WHERE source_rubric.position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

UPDATE trainee_assessments
SET position_id=@target_position_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT criterion_id,@target_position_id
FROM trainee_assessment_criterion_positions
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

DELETE FROM trainee_assessment_criterion_positions
WHERE position_id=@source_position_id AND @source_position_id IS NOT NULL;

DELETE score_rule FROM trainee_assessment_score_rules score_rule
JOIN trainee_assessment_rubrics rubric ON rubric.id=score_rule.rubric_id
WHERE rubric.position_id=@source_position_id AND @source_position_id IS NOT NULL;

DELETE FROM trainee_assessment_rubrics
WHERE position_id=@source_position_id AND @source_position_id IS NOT NULL;

UPDATE trainee_assessment_criteria
SET position_id=@target_position_id,department_id=@target_department_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

DELETE FROM positions
WHERE id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

COMMIT;

-- Gabungkan PM FRESHFOOD ke PM SM BUAH & SAYUR.
START TRANSACTION;

SET @source_position_id := (
  SELECT id FROM positions WHERE code='pm_freshfood' LIMIT 1
);
SET @target_position_id := (
  SELECT id FROM positions WHERE code='pm_sm_buah_sayur' LIMIT 1
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

UPDATE trainee_assessments
SET position_id=@target_position_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

UPDATE trainee_assessment_criteria
SET position_id=@target_position_id,department_id=@target_department_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT criterion_id,@target_position_id
FROM trainee_assessment_criterion_positions
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

DELETE FROM trainee_assessment_criterion_positions
WHERE position_id=@source_position_id AND @source_position_id IS NOT NULL;

UPDATE trainee_assessment_rubrics
SET position_id=@target_position_id,department_id=@target_department_id
WHERE position_id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

DELETE FROM positions
WHERE id=@source_position_id
  AND @source_position_id IS NOT NULL AND @target_position_id IS NOT NULL;

COMMIT;

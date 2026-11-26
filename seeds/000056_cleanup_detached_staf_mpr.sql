-- Bersihkan aturan STAF MPR yang sudah dilepas dan belum pernah dinilai.
START TRANSACTION;
SET @mpr := (SELECT id FROM positions WHERE code='staf_mpr');
DELETE s FROM trainee_assessment_score_rules s
JOIN trainee_assessment_rubrics r ON r.id=s.rubric_id
WHERE r.position_id=@mpr AND r.code IN ('group_18_1_2','group_18_3_4','group_18_5_6')
AND NOT EXISTS(SELECT 1 FROM trainee_assessment_criterion_positions cp WHERE cp.position_id=r.position_id AND cp.criterion_id=s.criterion_id)
AND NOT EXISTS(SELECT 1 FROM trainee_assessments a WHERE a.rubric_id=r.id);
DELETE FROM trainee_assessment_rubrics
WHERE position_id=@mpr AND code IN ('group_18_1_2','group_18_3_4','group_18_5_6')
AND NOT EXISTS(SELECT 1 FROM trainee_assessment_score_rules s WHERE s.rubric_id=trainee_assessment_rubrics.id)
AND NOT EXISTS(SELECT 1 FROM trainee_assessments a WHERE a.rubric_id=trainee_assessment_rubrics.id);
COMMIT;

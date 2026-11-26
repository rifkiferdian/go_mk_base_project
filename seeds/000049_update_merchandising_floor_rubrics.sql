-- Koreksi soal Trainee PM MERCHANDISING (FLOOR SM) ke panduan Food/Toiletris.
START TRANSACTION;
SET @floor_position := (SELECT id FROM positions WHERE code='pm_merchandising_floor_sm');

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,@floor_position FROM trainee_assessment_criteria c
JOIN positions p ON p.id=c.position_id
WHERE p.code='md_food' AND c.code LIKE 'md_food_toiletris_%' AND c.employment_status='trainee';

DELETE s FROM trainee_assessment_score_rules s
JOIN trainee_assessment_rubrics r ON r.id=s.rubric_id
JOIN trainee_assessment_criteria c ON c.id=s.criterion_id
WHERE r.position_id=@floor_position AND r.code IN ('pm_dept_store_1_2','pm_dept_store_3_4','pm_dept_store_5_6')
AND c.code LIKE 'pm_dept_store_%' AND c.employment_status='trainee';

DELETE cp FROM trainee_assessment_criterion_positions cp
JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id
WHERE cp.position_id=@floor_position AND c.code LIKE 'pm_dept_store_%' AND c.employment_status='trainee';

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT target.id,s.criterion_id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_rubrics target
JOIN positions p ON p.code='md_food'
JOIN trainee_assessment_rubrics source ON source.position_id=p.id AND source.month_from=target.month_from
JOIN trainee_assessment_score_rules s ON s.rubric_id=source.id
JOIN trainee_assessment_criteria c ON c.id=s.criterion_id
WHERE target.position_id=@floor_position AND target.code IN ('pm_dept_store_1_2','pm_dept_store_3_4','pm_dept_store_5_6')
AND source.code IN ('md_food_toiletris_1_2','md_food_toiletris_3_4','md_food_toiletris_5_6')
AND c.employment_status='trainee' AND c.code LIKE 'md_food_toiletris_%'
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

UPDATE trainee_assessment_rubrics SET code=CONCAT('floor_sm_',month_from,'_',month_to)
WHERE position_id=@floor_position AND code IN ('pm_dept_store_1_2','pm_dept_store_3_4','pm_dept_store_5_6');
COMMIT;

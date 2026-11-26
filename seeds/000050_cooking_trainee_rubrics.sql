-- Panduan PM KAFE untuk Trainee COOKING, memakai subaspek bersama WAITERS.
START TRANSACTION;

INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT p.department_id,p.id,CONCAT('cooking_',r.month_from,'_',r.month_to),CONCAT(p.name,' - Bulan ',r.month_from,'-',r.month_to),r.month_from,r.month_to,1
FROM positions p JOIN positions source ON source.code='waiters'
JOIN trainee_assessment_rubrics r ON r.position_id=source.id
WHERE p.code='cooking' AND r.code IN ('waiters_pm_kafe_1_2','waiters_pm_kafe_3_4','waiters_pm_kafe_5_6')
ON DUPLICATE KEY UPDATE name=VALUES(name),is_active=1;

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,p.id FROM trainee_assessment_criteria c CROSS JOIN positions p
WHERE p.code='cooking' AND c.code LIKE 'waiters_%' AND c.employment_status='trainee';

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT target.id,s.criterion_id,s.score,s.finding_min,s.finding_max,s.description
FROM positions p JOIN trainee_assessment_rubrics target ON target.position_id=p.id
JOIN positions source_position ON source_position.code='waiters'
JOIN trainee_assessment_rubrics source ON source.position_id=source_position.id AND source.month_from=target.month_from
JOIN trainee_assessment_score_rules s ON s.rubric_id=source.id
JOIN trainee_assessment_criteria c ON c.id=s.criterion_id
WHERE p.code='cooking' AND target.code IN ('cooking_1_2','cooking_3_4','cooking_5_6')
AND source.code IN ('waiters_pm_kafe_1_2','waiters_pm_kafe_3_4','waiters_pm_kafe_5_6')
AND c.code LIKE 'waiters_%' AND c.employment_status='trainee'
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

COMMIT;

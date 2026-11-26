-- Panduan Dept Store untuk jabatan lantai 2 dan merchandising floor.
START TRANSACTION;

INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT p.department_id,p.id,r.code,CONCAT(p.name,' - Bulan ',r.month_from,'-',r.month_to),r.month_from,r.month_to,1
FROM positions p
JOIN positions source ON source.code='pm_alat_tulis'
JOIN trainee_assessment_rubrics r ON r.position_id=source.id AND r.code IN ('pm_dept_store_1_2','pm_dept_store_3_4','pm_dept_store_5_6')
WHERE p.code IN ('pm_lantai_2_area_a','pm_lantai_2_area_b','pm_merchandising_floor_sm')
ON DUPLICATE KEY UPDATE name=VALUES(name),is_active=1;

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,p.id FROM trainee_assessment_criteria c CROSS JOIN positions p
WHERE c.code LIKE 'pm_dept_store_%'
AND p.code IN ('pm_lantai_2_area_a','pm_lantai_2_area_b','pm_merchandising_floor_sm');

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT target.id,s.criterion_id,s.score,s.finding_min,s.finding_max,s.description
FROM positions p
JOIN trainee_assessment_rubrics target ON target.position_id=p.id
JOIN positions source_position ON source_position.code='pm_alat_tulis'
JOIN trainee_assessment_rubrics source ON source.position_id=source_position.id AND source.code=target.code
JOIN trainee_assessment_score_rules s ON s.rubric_id=source.id
JOIN trainee_assessment_criteria c ON c.id=s.criterion_id AND c.code LIKE 'pm_dept_store_%'
WHERE p.code IN ('pm_lantai_2_area_a','pm_lantai_2_area_b','pm_merchandising_floor_sm')
AND target.code IN ('pm_dept_store_1_2','pm_dept_store_3_4','pm_dept_store_5_6')
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

COMMIT;

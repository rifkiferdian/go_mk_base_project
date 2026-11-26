-- STAF AKT : KAS BESAR memakai panduan HLPP bersama periode 1-6.
START TRANSACTION;
INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT department_id,id,'hlpp_1_6',CONCAT(name,' - HLPP Bulan 1-6'),1,6,1 FROM positions WHERE code='staf_akt_kas_besar'
ON DUPLICATE KEY UPDATE is_active=1;

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,p.id FROM trainee_assessment_criteria c CROSS JOIN positions p
WHERE p.code='staf_akt_kas_besar' AND c.code LIKE 'hlpp_umum_%' AND c.employment_status='post_retirement_daily_worker';

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT target.id,s.criterion_id,s.score,s.finding_min,s.finding_max,s.description
FROM positions p JOIN trainee_assessment_rubrics target ON target.position_id=p.id AND target.code='hlpp_1_6'
JOIN positions source_position ON source_position.code='input_edit_proses_pembelian'
JOIN trainee_assessment_rubrics source ON source.position_id=source_position.id AND source.code='hlpp_1_6'
JOIN trainee_assessment_score_rules s ON s.rubric_id=source.id
JOIN trainee_assessment_criteria c ON c.id=s.criterion_id
WHERE p.code='staf_akt_kas_besar' AND c.code LIKE 'hlpp_umum_%' AND c.employment_status='post_retirement_daily_worker'
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);
COMMIT;

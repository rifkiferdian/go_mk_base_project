-- Idempotent correction: one Gudang criterion set, three period-specific rule sets.
START TRANSACTION;
INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT r.id,c.id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_rubrics r
JOIN departments d ON d.id=r.department_id AND d.code='gudang'
JOIN trainee_assessment_criteria c ON c.department_id=d.id
JOIN (
 SELECT 'standard' kind,5 score,0 finding_min,0 finding_max,'tidak ada temuan' description
 UNION ALL SELECT 'standard',4,1,2,'1x - 2x temuan'
 UNION ALL SELECT 'standard',3,3,3,'3x temuan'
 UNION ALL SELECT 'standard',2,4,4,'4x temuan'
 UNION ALL SELECT 'standard',1,5,NULL,'≥ 5x temuan'
 UNION ALL SELECT 'early',5,0,2,'< 3x temuan'
 UNION ALL SELECT 'early',4,3,3,'3x temuan'
 UNION ALL SELECT 'early',3,4,4,'4x temuan'
 UNION ALL SELECT 'early',2,5,5,'5x temuan'
 UNION ALL SELECT 'early',1,6,NULL,'≥ 6x temuan'
 UNION ALL SELECT 'middle',5,0,1,'< 2x temuan'
 UNION ALL SELECT 'middle',4,2,2,'2x temuan'
 UNION ALL SELECT 'middle',3,3,3,'3x temuan'
 UNION ALL SELECT 'middle',2,4,4,'4x temuan'
 UNION ALL SELECT 'middle',1,5,NULL,'≥ 5x temuan'
 UNION ALL SELECT 'late',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late',4,1,1,'1x temuan'
 UNION ALL SELECT 'late',3,2,2,'2x temuan'
 UNION ALL SELECT 'late',2,3,3,'3x temuan'
 UNION ALL SELECT 'late',1,4,NULL,'≥ 4x temuan'
 UNION ALL SELECT 'late_social',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late_social',4,1,1,'1x temuan'
 UNION ALL SELECT 'late_social',3,2,2,'2x temuan'
 UNION ALL SELECT 'late_social',2,3,4,'3x - 4x temuan'
 UNION ALL SELECT 'late_social',1,5,NULL,'≥ 5x temuan'
) s ON s.kind=CASE
 WHEN r.month_from=5 AND c.code IN ('kerjasama_1','kerjasama_2','keinginan_belajar_1','keinginan_belajar_2') THEN 'late_social'
 WHEN r.month_from=5 THEN 'late'
 WHEN c.code IN ('layanan_1','layanan_2','etika_1') THEN 'standard'
 WHEN r.month_from=1 THEN 'early'
 ELSE 'middle' END
WHERE r.code IN ('gudang_1_2','gudang_3_4','gudang_5_plus')
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);
COMMIT;

INSERT INTO positions(job_level_id,department_id,code,name)
SELECT l.id,d.id,p.code,p.name
FROM departments d
JOIN job_levels l ON l.code='staff'
JOIN (
 SELECT 'cooking' AS code,'COOKING' AS name
 UNION ALL SELECT 'customer_service','CUSTOMER SERVICE'
 UNION ALL SELECT 'md_food','MD FOOD'
 UNION ALL SELECT 'md_toiletris','MD TOILETRIS'
 UNION ALL SELECT 'monitoring_cctv','MONITORING CCTV'
 UNION ALL SELECT 'waiters','WAITERS'
) p
WHERE d.code='operational'
ON DUPLICATE KEY UPDATE
 job_level_id=VALUES(job_level_id),
 department_id=VALUES(department_id),
 name=VALUES(name);

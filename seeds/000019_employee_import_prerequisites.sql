START TRANSACTION;

INSERT INTO stores(code,name,timezone,status)
VALUES('MK8','MK8','Asia/Jakarta','active')
ON DUPLICATE KEY UPDATE name=VALUES(name),timezone=VALUES(timezone),status='active',deleted_at=NULL;

INSERT INTO positions(job_level_id,department_id,code,name)
SELECT l.id,d.id,p.code,p.name
FROM departments d
JOIN job_levels l ON l.code='staff'
JOIN (
 SELECT 'operational_penataan_barang_picke' AS code,'PENATAAN BARANG DAN PICKE' AS name
 UNION ALL SELECT 'pm_bsn_kosmetik','PM BSN KOSMETIK'
 UNION ALL SELECT 'pm_freshfood','PM FRESHFOOD'
) p
WHERE d.code='operational'
ON DUPLICATE KEY UPDATE
 job_level_id=VALUES(job_level_id),
 department_id=VALUES(department_id),
 name=VALUES(name);

COMMIT;

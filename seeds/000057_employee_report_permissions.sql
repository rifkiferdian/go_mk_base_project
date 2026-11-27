START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES ('employee_reports.view','web','Melihat laporan penilaian per karyawan') ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id) SELECT DISTINCT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE p.name='employee_reports.view' AND (r.name='superadmin' OR EXISTS(SELECT 1 FROM role_permissions rp JOIN permissions existing ON existing.id=rp.permission_id WHERE rp.role_id=r.id AND existing.name IN ('team.view','trainee_assessment_reports.view')));
COMMIT;

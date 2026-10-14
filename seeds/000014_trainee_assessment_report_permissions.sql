INSERT INTO permissions(name,guard_name,description) VALUES
('trainee_assessment_reports.view','web','Melihat laporan penilaian trainee')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND p.name='trainee_assessment_reports.view';

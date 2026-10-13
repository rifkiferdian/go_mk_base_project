START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('assessment_criteria.view','web','Melihat master subaspek penilaian trainee'),
('assessment_criteria.create','web','Membuat subaspek penilaian trainee'),
('assessment_criteria.update','web','Memperbarui subaspek penilaian trainee'),
('assessment_criteria.delete','web','Menghapus subaspek penilaian trainee')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'assessment_criteria.%';
COMMIT;

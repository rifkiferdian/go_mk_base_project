START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('assessment_aspects.view','web','Melihat master aspek penilaian trainee'),
('assessment_aspects.create','web','Membuat aspek penilaian trainee'),
('assessment_aspects.update','web','Memperbarui aspek penilaian trainee'),
('assessment_aspects.delete','web','Menghapus aspek penilaian trainee')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'assessment_aspects.%';
COMMIT;

START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('assessment_result_categories.view','web','Melihat master kategori hasil penilaian'),
('assessment_result_categories.create','web','Membuat kategori hasil penilaian'),
('assessment_result_categories.update','web','Memperbarui kategori hasil penilaian'),
('assessment_result_categories.delete','web','Menghapus kategori hasil penilaian')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'assessment_result_categories.%';
COMMIT;

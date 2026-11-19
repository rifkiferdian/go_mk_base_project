START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('assessment_rubrics.view','web','Melihat periode rubrik'),('assessment_rubrics.create','web','Membuat periode rubrik'),('assessment_rubrics.update','web','Memperbarui periode rubrik'),('assessment_rubrics.delete','web','Menghapus periode rubrik')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id) SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.name='superadmin' AND p.name LIKE 'assessment_rubrics.%';
COMMIT;

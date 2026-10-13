INSERT INTO permissions(name,guard_name,description) VALUES
('trainee_assessments.view','web','Melihat daftar penilaian trainee'),
('trainee_assessments.create','web','Membuat penilaian trainee'),
('trainee_assessments.update','web','Mengubah dan mengajukan penilaian trainee'),
('trainee_assessments.delete','web','Menghapus draft penilaian trainee')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND p.name LIKE 'trainee_assessments.%';

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name IN ('manager','supervisor') AND p.name IN ('trainee_assessments.view','trainee_assessments.create','trainee_assessments.update');

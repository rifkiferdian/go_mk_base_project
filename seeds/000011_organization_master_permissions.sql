START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('job_levels.view','web','Melihat master tingkatan jabatan'),('job_levels.create','web','Membuat tingkatan jabatan'),('job_levels.update','web','Memperbarui tingkatan jabatan'),('job_levels.delete','web','Menghapus tingkatan jabatan'),
('positions.view','web','Melihat master jabatan'),('positions.create','web','Membuat jabatan'),('positions.update','web','Memperbarui jabatan'),('positions.delete','web','Menghapus jabatan'),
('departments.view','web','Melihat master bagian'),('departments.create','web','Membuat bagian'),('departments.update','web','Memperbarui bagian'),('departments.delete','web','Menghapus bagian')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.name='superadmin' AND r.guard_name='web' AND (p.name LIKE 'job_levels.%' OR p.name LIKE 'positions.%' OR p.name LIKE 'departments.%');
COMMIT;

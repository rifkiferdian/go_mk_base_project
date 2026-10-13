INSERT INTO permissions(name,guard_name,description) VALUES
('employees.view','web','Melihat data seluruh karyawan'),
('employees.create','web','Menambah karyawan'),
('employees.update','web','Mengubah data karyawan'),
('employees.delete','web','Menghapus karyawan'),
('employee_profiles.view','web','Melihat biodata karyawan'),
('employee_profiles.create','web','Menambah biodata karyawan'),
('employee_profiles.update','web','Mengubah biodata karyawan'),
('employee_profiles.delete','web','Menghapus biodata karyawan'),
('employee_assignments.view','web','Melihat riwayat penempatan karyawan'),
('employee_assignments.create','web','Menambah penempatan karyawan'),
('employee_assignments.update','web','Mengubah penempatan karyawan'),
('employee_assignments.delete','web','Menghapus penempatan karyawan'),
('employee_reporting_lines.view','web','Melihat hubungan atasan dan bawahan'),
('employee_reporting_lines.create','web','Menambah hubungan atasan dan bawahan'),
('employee_reporting_lines.update','web','Mengubah hubungan atasan dan bawahan'),
('employee_reporting_lines.delete','web','Menghapus hubungan atasan dan bawahan')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND (p.name LIKE 'employees.%' OR p.name LIKE 'employee_profiles.%' OR p.name LIKE 'employee_assignments.%' OR p.name LIKE 'employee_reporting_lines.%');

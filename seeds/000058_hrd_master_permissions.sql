START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('hrd_categories.view','web','Melihat kategori potongan HRD'),
('hrd_categories.manage','web','Mengelola kategori potongan HRD'),
('hrd_rules.view','web','Melihat aturan potongan HRD'),
('hrd_rules.manage','web','Mengelola aturan potongan HRD')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id) SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.name='superadmin' AND p.name IN ('hrd_categories.view','hrd_categories.manage','hrd_rules.view','hrd_rules.manage');
COMMIT;

START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES('team.view','web','Melihat daftar bawahan aktif')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name IN ('superadmin','supervisor','manager') AND p.name='team.view';
COMMIT;

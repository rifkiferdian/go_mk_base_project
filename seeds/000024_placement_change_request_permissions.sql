START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('placement_change_requests.view','web','Melihat pengajuan perubahan penempatan'),
('placement_change_requests.create','web','Mengajukan perubahan penempatan bawahan'),
('placement_change_requests.approve','web','Menyetujui atau menolak perubahan penempatan')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE (r.name='superadmin') OR (r.name='supervisor' AND p.name IN ('placement_change_requests.view','placement_change_requests.create')) OR (r.name='manager' AND p.name IN ('placement_change_requests.view','placement_change_requests.approve'));
COMMIT;

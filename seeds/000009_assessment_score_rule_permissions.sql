START TRANSACTION;
INSERT INTO permissions(name,guard_name,description) VALUES
('assessment_score_rules.view','web','Melihat master aturan skor penilaian trainee'),
('assessment_score_rules.create','web','Membuat aturan skor penilaian trainee'),
('assessment_score_rules.update','web','Memperbarui aturan skor penilaian trainee'),
('assessment_score_rules.delete','web','Menghapus aturan skor penilaian trainee')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT IGNORE INTO role_permissions(role_id,permission_id)
SELECT r.id,p.id FROM roles r CROSS JOIN permissions p
WHERE r.name='superadmin' AND r.guard_name='web' AND p.name LIKE 'assessment_score_rules.%';
COMMIT;

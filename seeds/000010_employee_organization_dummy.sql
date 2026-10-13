-- Hanya development. Seluruh identitas di bawah adalah contoh fiktif.
START TRANSACTION;
INSERT INTO job_levels(code,name,rank_order) VALUES
('director','Direktur',1),('area_manager','Area Manager',2),('store_manager','Store Manager',3),
('assistant_manager','Assistant Store Manager',4),('supervisor','Supervisor',5),('staff','Staff / Trainee',6)
ON DUPLICATE KEY UPDATE code=VALUES(code);
INSERT INTO positions(job_level_id,code,name)
SELECT l.id,p.code,p.name FROM job_levels l JOIN (
 SELECT 'store_manager' level_code,'store_manager' code,'Store Manager' name
 UNION ALL SELECT 'supervisor','spv_gudang','Supervisor Gudang'
 UNION ALL SELECT 'supervisor','spv_kasir','Supervisor Kasir'
 UNION ALL SELECT 'staff','staff_gudang','Staff Gudang'
 UNION ALL SELECT 'staff','trainee_gudang','Trainee Gudang'
 UNION ALL SELECT 'staff','trainee_kasir','Trainee Kasir'
) p ON p.level_code=l.code ON DUPLICATE KEY UPDATE code=VALUES(code);
INSERT INTO departments(code,name) VALUES ('operasional','Operasional'),('gudang','Gudang'),('kasir','Kasir'),('hr','Human Resources')
ON DUPLICATE KEY UPDATE code=VALUES(code);
INSERT INTO employees(employee_number,name,email,phone) VALUES
('DUMMY-MGR-001','Budi Contoh','budi@example.invalid','080000000001'),
('DUMMY-SPV-001','Sari Contoh','sari@example.invalid','080000000002'),
('DUMMY-SPV-002','Dewi Contoh','dewi@example.invalid','080000000003'),
('DUMMY-STF-001','Andi Contoh','andi@example.invalid','080000000004'),
('DUMMY-TRN-001','Rina Contoh','rina@example.invalid','080000000005'),
('DUMMY-TRN-002','Dimas Contoh','dimas@example.invalid','080000000006')
ON DUPLICATE KEY UPDATE employee_number=VALUES(employee_number);
INSERT INTO employee_profiles(employee_id,birth_place,birth_date,gender,address,joined_date,employment_status,emergency_contact_name,emergency_contact_phone,emergency_contact_relationship)
SELECT e.id,'Yogyakarta','1998-01-15',CASE WHEN e.employee_number IN ('DUMMY-SPV-001','DUMMY-SPV-002','DUMMY-TRN-001') THEN 'female' ELSE 'male' END,
 'Alamat contoh fiktif, Yogyakarta',CASE WHEN e.employee_number LIKE 'DUMMY-TRN-%' THEN '2026-07-01' ELSE '2024-01-01' END,
 CASE WHEN e.employee_number LIKE 'DUMMY-TRN-%' THEN 'trainee' ELSE 'permanent' END,'Keluarga Contoh','080000000099','Keluarga'
FROM employees e WHERE e.employee_number IN ('DUMMY-MGR-001','DUMMY-SPV-001','DUMMY-SPV-002','DUMMY-STF-001','DUMMY-TRN-001','DUMMY-TRN-002')
ON DUPLICATE KEY UPDATE employee_id=VALUES(employee_id);
INSERT INTO employee_assignments(employee_id,position_id,department_id,store_id,started_at)
SELECT e.id,p.id,d.id,s.id,CASE WHEN e.employee_number LIKE 'DUMMY-TRN-%' THEN '2026-07-01' ELSE '2024-01-01' END
FROM employees e JOIN (
 SELECT 'DUMMY-MGR-001' nik,'store_manager' position_code,'operasional' department_code
 UNION ALL SELECT 'DUMMY-SPV-001','spv_gudang','gudang'
 UNION ALL SELECT 'DUMMY-SPV-002','spv_kasir','kasir'
 UNION ALL SELECT 'DUMMY-STF-001','staff_gudang','gudang'
 UNION ALL SELECT 'DUMMY-TRN-001','trainee_gudang','gudang'
 UNION ALL SELECT 'DUMMY-TRN-002','trainee_kasir','kasir'
) x ON x.nik=e.employee_number JOIN positions p ON p.code=x.position_code
JOIN departments d ON d.code=x.department_code JOIN stores s ON s.code='MK5'
ON DUPLICATE KEY UPDATE employee_id=VALUES(employee_id);
INSERT INTO employee_reporting_lines(employee_assignment_id,supervisor_assignment_id,started_at)
SELECT child.id,parent.id,child.started_at FROM (
 SELECT 'DUMMY-SPV-001' child_nik,'DUMMY-MGR-001' parent_nik
 UNION ALL SELECT 'DUMMY-SPV-002','DUMMY-MGR-001'
 UNION ALL SELECT 'DUMMY-STF-001','DUMMY-SPV-001'
 UNION ALL SELECT 'DUMMY-TRN-001','DUMMY-SPV-001'
 UNION ALL SELECT 'DUMMY-TRN-002','DUMMY-SPV-002'
) x JOIN employees ce ON ce.employee_number=x.child_nik JOIN employee_assignments child ON child.employee_id=ce.id AND child.ended_at IS NULL
JOIN employees pe ON pe.employee_number=x.parent_nik JOIN employee_assignments parent ON parent.employee_id=pe.id AND parent.ended_at IS NULL
ON DUPLICATE KEY UPDATE employee_assignment_id=VALUES(employee_assignment_id);

-- Akun contoh hanya manager/SPV, password development qweqwe.
INSERT INTO users(employee_id,name,username,email,password_hash,status)
SELECT e.id,e.name,CASE e.employee_number WHEN 'DUMMY-MGR-001' THEN 'dummy.manager' WHEN 'DUMMY-SPV-001' THEN 'dummy.spv.gudang' ELSE 'dummy.spv.kasir' END,e.email,
 '$2y$10$lwkQLoTL4oDoSPj6oGHuL.bhCFZzZE3hfwul5dw6f738FrThFhov6','active'
FROM employees e WHERE e.employee_number IN ('DUMMY-MGR-001','DUMMY-SPV-001','DUMMY-SPV-002')
ON DUPLICATE KEY UPDATE username=VALUES(username);
INSERT INTO roles(name,guard_name,description) VALUES ('manager','web','Manager retail'),('supervisor','web','Supervisor retail')
ON DUPLICATE KEY UPDATE name=VALUES(name);
INSERT INTO user_stores(user_id,store_id,is_default,status,joined_at)
SELECT u.id,s.id,1,'active',CURRENT_TIMESTAMP(6) FROM users u JOIN employees e ON e.id=u.employee_id JOIN stores s ON s.code='MK5'
WHERE e.employee_number IN ('DUMMY-MGR-001','DUMMY-SPV-001','DUMMY-SPV-002')
ON DUPLICATE KEY UPDATE user_id=VALUES(user_id);
INSERT IGNORE INTO user_store_roles(user_store_id,role_id)
SELECT us.id,r.id FROM user_stores us JOIN users u ON u.id=us.user_id JOIN employees e ON e.id=u.employee_id
JOIN roles r ON r.guard_name='web' AND r.name=CASE WHEN e.employee_number='DUMMY-MGR-001' THEN 'manager' ELSE 'supervisor' END
WHERE e.employee_number IN ('DUMMY-MGR-001','DUMMY-SPV-001','DUMMY-SPV-002');
COMMIT;

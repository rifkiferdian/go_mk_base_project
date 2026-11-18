START TRANSACTION;

CREATE TEMPORARY TABLE temporary_supervisors (
 employee_number VARCHAR(50) NOT NULL,
 employee_name VARCHAR(150) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO temporary_supervisors(employee_number,employee_name) VALUES
('SPV01','HAZMI'),
('SPV02','APRILIA'),
('SPV03','RIRI'),
('SPV04','APRIL'),
('SPV05','RISTI'),
('SPV06','GUSTOMI'),
('SPV07','IBNU'),
('SPV08','ANNISA'),
('SPV09','DEWI'),
('SPV10','ABY'),
('SPV11','ANDREAS'),
('SPV12','MISKAN'),
('SPV13','SIGIT'),
('SPV14','EKY'),
('SPV15','LIRIS'),
('SPV16','ANANG'),
('SPV17','MARULI'),
('SPV18','NURITA'),
('SPV19','FIKRI'),
('SPV20','PRAMONO'),
('SPV21','SIVA'),
('SPV22','UNGGARA'),
('SPV23','YATIN'),
('SPV24','HARYANTO'),
('SPV25','ARFIANI'),
('SPV26','SURANA'),
('SPV27','WIWIT'),
('SPV28','LIA'),
('SPV29','EKO'),
('SPV30','HESA');

INSERT INTO employees(employee_number,name,status)
SELECT employee_number,employee_name,'active'
FROM temporary_supervisors
ON DUPLICATE KEY UPDATE name=VALUES(name),status='active';

INSERT INTO employee_profiles(employee_id,employment_status)
SELECT e.id,'permanent'
FROM temporary_supervisors t
JOIN employees e ON e.employee_number=t.employee_number
ON DUPLICATE KEY UPDATE employment_status=VALUES(employment_status);

DROP TEMPORARY TABLE temporary_supervisors;
COMMIT;

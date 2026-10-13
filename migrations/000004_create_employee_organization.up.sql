CREATE TABLE job_levels (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 code VARCHAR(50) NOT NULL UNIQUE,
 name VARCHAR(100) NOT NULL,
 rank_order SMALLINT UNSIGNED NOT NULL,
 CHECK (rank_order > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE positions (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 job_level_id BIGINT UNSIGNED NOT NULL,
 code VARCHAR(50) NOT NULL UNIQUE,
 name VARCHAR(150) NOT NULL,
 FOREIGN KEY (job_level_id) REFERENCES job_levels(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE departments (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 code VARCHAR(50) NOT NULL UNIQUE,
 name VARCHAR(150) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE employees (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 employee_number VARCHAR(50) NOT NULL UNIQUE,
 name VARCHAR(150) NOT NULL,
 email VARCHAR(191) NULL,
 phone VARCHAR(30) NULL,
 status ENUM('active','inactive') NOT NULL DEFAULT 'active',
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE employee_profiles (
 employee_id BIGINT UNSIGNED PRIMARY KEY,
 birth_place VARCHAR(100) NULL,
 birth_date DATE NULL,
 gender ENUM('male','female') NULL,
 address TEXT NULL,
 joined_date DATE NULL,
 employment_status ENUM('trainee','probation','contract','permanent') NOT NULL DEFAULT 'trainee',
 photo_path VARCHAR(255) NULL,
 emergency_contact_name VARCHAR(150) NULL,
 emergency_contact_phone VARCHAR(30) NULL,
 emergency_contact_relationship VARCHAR(50) NULL,
 updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
 FOREIGN KEY (employee_id) REFERENCES employees(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE employee_assignments (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 employee_id BIGINT UNSIGNED NOT NULL,
 position_id BIGINT UNSIGNED NOT NULL,
 department_id BIGINT UNSIGNED NOT NULL,
 store_id BIGINT UNSIGNED NOT NULL,
 started_at DATE NOT NULL,
 ended_at DATE NULL,
 current_employee_id BIGINT UNSIGNED GENERATED ALWAYS AS (CASE WHEN ended_at IS NULL THEN employee_id ELSE NULL END) STORED,
 UNIQUE KEY employee_current_assignment_unique (current_employee_id),
 UNIQUE KEY employee_assignment_start_unique (employee_id,started_at),
 KEY assignment_store_department (store_id,department_id),
 CHECK (ended_at IS NULL OR ended_at >= started_at),
 FOREIGN KEY (employee_id) REFERENCES employees(id),
 FOREIGN KEY (position_id) REFERENCES positions(id),
 FOREIGN KEY (department_id) REFERENCES departments(id),
 FOREIGN KEY (store_id) REFERENCES stores(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE employee_reporting_lines (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 employee_assignment_id BIGINT UNSIGNED NOT NULL,
 supervisor_assignment_id BIGINT UNSIGNED NOT NULL,
 started_at DATE NOT NULL,
 ended_at DATE NULL,
 current_assignment_id BIGINT UNSIGNED GENERATED ALWAYS AS (CASE WHEN ended_at IS NULL THEN employee_assignment_id ELSE NULL END) STORED,
 UNIQUE KEY employee_current_supervisor_unique (current_assignment_id),
 UNIQUE KEY reporting_start_unique (employee_assignment_id,started_at),
 KEY reporting_supervisor_index (supervisor_assignment_id),
 CHECK (employee_assignment_id <> supervisor_assignment_id),
 CHECK (ended_at IS NULL OR ended_at >= started_at),
 FOREIGN KEY (employee_assignment_id) REFERENCES employee_assignments(id),
 FOREIGN KEY (supervisor_assignment_id) REFERENCES employee_assignments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE users
 ADD COLUMN employee_id BIGINT UNSIGNED NULL,
 ADD UNIQUE KEY users_employee_id_unique (employee_id),
 ADD CONSTRAINT users_employee_fk FOREIGN KEY (employee_id) REFERENCES employees(id);

-- Konversi karyawan yang telah muncul dalam penilaian lama, jika ada.
INSERT INTO employees(employee_number,name,email,phone,status)
SELECT COALESCE(NULLIF(u.employee_number,''),CONCAT('LEGACY-USER-',u.id)),u.name,u.email,u.phone,
 CASE WHEN u.status='active' THEN 'active' ELSE 'inactive' END
FROM users u WHERE EXISTS(SELECT 1 FROM trainee_assessments a WHERE a.trainee_user_id=u.id)
ON DUPLICATE KEY UPDATE employee_number=VALUES(employee_number);
UPDATE users u JOIN employees e ON e.employee_number=COALESCE(NULLIF(u.employee_number,''),CONCAT('LEGACY-USER-',u.id))
SET u.employee_id=e.id WHERE u.employee_id IS NULL;

-- Kolom lama dipertahankan nullable untuk kompatibilitas dan rollback.
ALTER TABLE trainee_assessments
 MODIFY trainee_user_id BIGINT UNSIGNED NULL,
 ADD COLUMN trainee_employee_id BIGINT UNSIGNED NULL,
 ADD COLUMN trainee_assignment_id BIGINT UNSIGNED NULL,
 ADD COLUMN assessor_assignment_id BIGINT UNSIGNED NULL,
 ADD COLUMN trainee_position_name VARCHAR(150) NULL,
 ADD COLUMN assessor_position_name VARCHAR(150) NULL,
 ADD CONSTRAINT assessment_employee_fk FOREIGN KEY (trainee_employee_id) REFERENCES employees(id),
 ADD CONSTRAINT assessment_trainee_assignment_fk FOREIGN KEY (trainee_assignment_id) REFERENCES employee_assignments(id),
 ADD CONSTRAINT assessment_assessor_assignment_fk FOREIGN KEY (assessor_assignment_id) REFERENCES employee_assignments(id);
UPDATE trainee_assessments a JOIN users u ON u.id=a.trainee_user_id SET a.trainee_employee_id=u.employee_id;
ALTER TABLE trainee_assessments
 MODIFY trainee_employee_id BIGINT UNSIGNED NOT NULL,
 ADD UNIQUE KEY assessment_employee_month_unique (trainee_employee_id,training_month);

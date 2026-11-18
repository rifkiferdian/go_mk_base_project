ALTER TABLE employee_profiles
 MODIFY employment_status ENUM('trainee','probation','contract','permanent','daily_worker','post_retirement_daily_worker') NOT NULL DEFAULT 'trainee';

UPDATE employee_profiles p
JOIN employees e ON e.id=p.employee_id
SET p.employment_status='post_retirement_daily_worker'
WHERE e.employee_number IN ('960012','960007','960008','910001','HLPP-KUPON','HLPP-YUNAN-SETIYAWARNI','HLPP-HARIS-SERA-PASARUA','HLPP-PRASETYO-BUDI-SANTOSO','HLPP-SUGENG-RETNO');

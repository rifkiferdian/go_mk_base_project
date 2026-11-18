UPDATE employee_profiles SET employment_status='contract'
WHERE employment_status IN ('daily_worker','post_retirement_daily_worker');

ALTER TABLE employee_profiles
 MODIFY employment_status ENUM('trainee','probation','contract','permanent') NOT NULL DEFAULT 'trainee';

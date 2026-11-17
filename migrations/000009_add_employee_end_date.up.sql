ALTER TABLE employee_profiles
 ADD COLUMN employment_ended_at DATE NULL AFTER joined_date,
 ADD CONSTRAINT employee_profile_employment_period_check
 CHECK (employment_ended_at IS NULL OR joined_date IS NULL OR employment_ended_at >= joined_date);

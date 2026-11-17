ALTER TABLE positions ADD COLUMN department_id BIGINT UNSIGNED NULL AFTER job_level_id;

UPDATE positions p
JOIN (
 SELECT position_id,MIN(department_id) AS department_id
 FROM employee_assignments
 GROUP BY position_id
) a ON a.position_id=p.id
SET p.department_id=a.department_id;

ALTER TABLE positions
 ADD KEY positions_department_index (department_id),
 ADD CONSTRAINT positions_department_fk FOREIGN KEY (department_id) REFERENCES departments(id);

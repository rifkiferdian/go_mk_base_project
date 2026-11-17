ALTER TABLE positions DROP FOREIGN KEY positions_department_fk;
ALTER TABLE positions DROP INDEX positions_department_index;
ALTER TABLE positions DROP COLUMN department_id;

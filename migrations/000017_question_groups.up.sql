CREATE TABLE assessment_question_groups (
 id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
 code VARCHAR(80) NOT NULL UNIQUE,
 name VARCHAR(150) NOT NULL,
 description VARCHAR(1000) NOT NULL DEFAULT ''
);
ALTER TABLE trainee_assessment_criteria ADD COLUMN question_group_id BIGINT UNSIGNED NULL,
 ADD CONSTRAINT criterion_question_group_fk FOREIGN KEY(question_group_id) REFERENCES assessment_question_groups(id);

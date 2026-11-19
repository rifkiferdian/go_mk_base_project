CREATE TABLE trainee_assessment_result_categories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    minimum_score DECIMAL(4,2) NOT NULL,
    maximum_score DECIMAL(4,2) NOT NULL,
    deduction_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    sort_order SMALLINT UNSIGNED NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    UNIQUE KEY result_category_name_unique (name),
    CONSTRAINT result_category_score_check CHECK (minimum_score >= 1 AND maximum_score <= 5 AND maximum_score >= minimum_score),
    CONSTRAINT result_category_deduction_check CHECK (deduction_percentage BETWEEN 0 AND 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO trainee_assessment_result_categories(name,minimum_score,maximum_score,deduction_percentage,sort_order) VALUES
('SANGAT KURANG SEKALI',1.00,1.80,100,1),
('KURANG SEKALI',1.81,2.60,75,2),
('KURANG',2.61,3.40,50,3),
('MENDEKATI STANDAR',3.41,4.20,25,4),
('SESUAI STANDAR',4.21,5.00,0,5);

ALTER TABLE trainee_assessments
    ADD COLUMN result_category_name VARCHAR(100) NULL AFTER status,
    ADD COLUMN result_deduction_percentage DECIMAL(5,2) NULL AFTER result_category_name;

UPDATE trainee_assessments a
JOIN trainee_assessment_results v ON v.assessment_id=a.id
JOIN trainee_assessment_result_categories c ON v.average_score BETWEEN c.minimum_score AND c.maximum_score
SET a.result_category_name=c.name,a.result_deduction_percentage=c.deduction_percentage
WHERE a.status='final';

CREATE OR REPLACE VIEW trainee_assessment_results AS
SELECT totals.assessment_id,totals.criterion_count,totals.scored_count,totals.total_weight,totals.total_score,totals.average_score,
       COALESCE(totals.result_category_name,(SELECT c.name FROM trainee_assessment_result_categories c WHERE c.is_active=1 AND totals.average_score BETWEEN c.minimum_score AND c.maximum_score ORDER BY c.sort_order LIMIT 1)) AS rating,
       COALESCE(totals.result_deduction_percentage,(SELECT c.deduction_percentage FROM trainee_assessment_result_categories c WHERE c.is_active=1 AND totals.average_score BETWEEN c.minimum_score AND c.maximum_score ORDER BY c.sort_order LIMIT 1)) AS deduction_percentage
FROM (
    SELECT a.id AS assessment_id,a.result_category_name,a.result_deduction_percentage,COUNT(d.criterion_id) AS criterion_count,
           COUNT(d.score) AS scored_count,SUM(d.weight) AS total_weight,SUM(d.weight*d.score) AS total_score,
           CASE WHEN COUNT(d.criterion_id)>0 AND COUNT(d.score)=COUNT(d.criterion_id) THEN ROUND(SUM(d.weight*d.score)/NULLIF(SUM(d.weight),0),2) ELSE NULL END AS average_score
    FROM trainee_assessments a LEFT JOIN trainee_assessment_details d ON d.assessment_id=a.id
    GROUP BY a.id,a.result_category_name,a.result_deduction_percentage
) totals;

CREATE OR REPLACE VIEW trainee_assessment_results AS
SELECT totals.*,
    CASE WHEN average_score <= 1.80 THEN 'Sangat Kurang Sekali' WHEN average_score <= 2.60 THEN 'Kurang Sekali' WHEN average_score <= 3.40 THEN 'Kurang' WHEN average_score <= 4.20 THEN 'Mendekati Standar' WHEN average_score <= 5.00 THEN 'Sesuai Standar' ELSE NULL END AS rating
FROM (SELECT a.id AS assessment_id,COUNT(d.criterion_id) AS criterion_count,COUNT(d.score) AS scored_count,SUM(d.weight) AS total_weight,SUM(d.weight*d.score) AS total_score,CASE WHEN COUNT(d.criterion_id)>0 AND COUNT(d.score)=COUNT(d.criterion_id) THEN ROUND(SUM(d.weight*d.score)/NULLIF(SUM(d.weight),0),2) ELSE NULL END AS average_score FROM trainee_assessments a LEFT JOIN trainee_assessment_details d ON d.assessment_id=a.id GROUP BY a.id) totals;
ALTER TABLE trainee_assessments DROP COLUMN result_deduction_percentage,DROP COLUMN result_category_name;
DROP TABLE trainee_assessment_result_categories;

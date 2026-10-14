package assessmentreport

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Summary(ctx context.Context) (Summary, error) {
	var result Summary
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(DISTINCT a.trainee_employee_id),COUNT(DISTINCT a.store_id),COALESCE(AVG(v.average_score),0),COALESCE(MAX(v.average_score),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.status='final'`).Scan(
		&result.AssessmentCount, &result.TraineeCount, &result.StoreCount, &result.AverageScore, &result.HighestScore)
	return result, err
}

func (r *Repository) Distribution(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT v.rating,COUNT(*) FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.status='final' AND v.rating IS NOT NULL GROUP BY v.rating`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]int)
	for rows.Next() {
		var label string
		var count int
		if err := rows.Scan(&label, &count); err != nil {
			return nil, err
		}
		result[label] = count
	}
	return result, rows.Err()
}

func (r *Repository) performance(ctx context.Context, query string) ([]Performance, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Performance
	for rows.Next() {
		var item Performance
		if err := rows.Scan(&item.Name, &item.Count, &item.Average); err != nil {
			return nil, err
		}
		item.Percentage = item.Average * 20
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Aspects(ctx context.Context) ([]Performance, error) {
	return r.performance(ctx, `SELECT d.aspect_name,COUNT(DISTINCT a.id),COALESCE(ROUND(SUM(d.weight*d.score)/NULLIF(SUM(d.weight),0),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_details d ON d.assessment_id=a.id
		WHERE a.status='final' AND d.score IS NOT NULL GROUP BY d.aspect_name ORDER BY MIN(d.criterion_id)`)
}

func (r *Repository) Stores(ctx context.Context) ([]Performance, error) {
	return r.performance(ctx, `SELECT a.store_name,COUNT(*),COALESCE(ROUND(AVG(v.average_score),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id
		WHERE a.status='final' GROUP BY a.store_name ORDER BY AVG(v.average_score) DESC,a.store_name LIMIT 8`)
}

func (r *Repository) Departments(ctx context.Context) ([]Performance, error) {
	return r.performance(ctx, `SELECT a.department_name,COUNT(*),COALESCE(ROUND(AVG(v.average_score),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id
		WHERE a.status='final' GROUP BY a.department_name ORDER BY AVG(v.average_score) DESC,a.department_name LIMIT 8`)
}

func (r *Repository) Recent(ctx context.Context) ([]RecentResult, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT a.employee_name,a.employee_number,COALESCE(a.trainee_position_name,''),a.department_name,a.store_name,a.training_month,DATE_FORMAT(a.assessment_date,'%d-%m-%Y'),a.assessor_name,COALESCE(v.average_score,0),COALESCE(v.rating,'')
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.status='final' ORDER BY a.assessment_date DESC,a.id DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []RecentResult
	for rows.Next() {
		var item RecentResult
		if err := rows.Scan(&item.EmployeeName, &item.EmployeeNumber, &item.PositionName, &item.DepartmentName, &item.StoreName, &item.TrainingMonth, &item.AssessmentDate, &item.AssessorName, &item.AverageScore, &item.Rating); err != nil {
			return nil, fmt.Errorf("scan recent result: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

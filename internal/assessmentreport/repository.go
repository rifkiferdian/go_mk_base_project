package assessmentreport

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func filterArgs(filter Filter) []any {
	return []any{filter.DepartmentID, filter.DepartmentID, filter.RubricID, filter.RubricID}
}

const reportFilter = ` AND (?=0 OR a.department_id=?) AND (?=0 OR a.rubric_id=?)`

func (r *Repository) Summary(ctx context.Context, filter Filter) (Summary, error) {
	var result Summary
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(DISTINCT a.trainee_employee_id),COUNT(DISTINCT a.store_id),COALESCE(AVG(v.average_score),0),COALESCE(MAX(v.average_score),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.status='final'`+reportFilter, filterArgs(filter)...).Scan(
		&result.AssessmentCount, &result.TraineeCount, &result.StoreCount, &result.AverageScore, &result.HighestScore)
	return result, err
}

func (r *Repository) Distribution(ctx context.Context, filter Filter) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT v.rating,COUNT(*) FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id WHERE a.status='final' AND v.rating IS NOT NULL`+reportFilter+` GROUP BY v.rating`, filterArgs(filter)...)
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

func (r *Repository) performance(ctx context.Context, query string, filter Filter) ([]Performance, error) {
	rows, err := r.db.QueryContext(ctx, query, filterArgs(filter)...)
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

func (r *Repository) Aspects(ctx context.Context, filter Filter) ([]Performance, error) {
	return r.performance(ctx, `SELECT d.aspect_name,COUNT(DISTINCT a.id),COALESCE(ROUND(SUM(d.weight*d.score)/NULLIF(SUM(d.weight),0),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_details d ON d.assessment_id=a.id
		WHERE a.status='final' AND d.score IS NOT NULL`+reportFilter+` GROUP BY d.aspect_name ORDER BY MIN(d.criterion_id)`, filter)
}

func (r *Repository) Stores(ctx context.Context, filter Filter) ([]Performance, error) {
	return r.performance(ctx, `SELECT a.store_name,COUNT(*),COALESCE(ROUND(AVG(v.average_score),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id
		WHERE a.status='final'`+reportFilter+` GROUP BY a.store_name ORDER BY AVG(v.average_score) DESC,a.store_name LIMIT 8`, filter)
}

func (r *Repository) Departments(ctx context.Context, filter Filter) ([]Performance, error) {
	return r.performance(ctx, `SELECT a.department_name,COUNT(*),COALESCE(ROUND(AVG(v.average_score),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id
		WHERE a.status='final'`+reportFilter+` GROUP BY a.department_name ORDER BY AVG(v.average_score) DESC,a.department_name LIMIT 8`, filter)
}

func (r *Repository) Periods(ctx context.Context, filter Filter) ([]Performance, error) {
	return r.performance(ctx, `SELECT COALESCE(b.name,CONCAT('Bulan ',a.training_month)),COUNT(*),COALESCE(ROUND(AVG(v.average_score),2),0)
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id LEFT JOIN trainee_assessment_rubrics b ON b.id=a.rubric_id
		WHERE a.status='final'`+reportFilter+` GROUP BY COALESCE(b.name,CONCAT('Bulan ',a.training_month)),COALESCE(b.month_from,a.training_month) ORDER BY COALESCE(b.month_from,a.training_month)`, filter)
}

func (r *Repository) Recent(ctx context.Context, filter Filter) ([]RecentResult, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT a.employee_name,a.employee_number,COALESCE(a.trainee_position_name,''),a.department_name,a.store_name,a.training_month,COALESCE(b.name,CONCAT('Bulan ',a.training_month)),DATE_FORMAT(a.assessment_date,'%d-%m-%Y'),a.assessor_name,COALESCE(v.average_score,0),COALESCE(v.rating,'')
		FROM trainee_assessments a JOIN trainee_assessment_results v ON v.assessment_id=a.id LEFT JOIN trainee_assessment_rubrics b ON b.id=a.rubric_id WHERE a.status='final'`+reportFilter+` ORDER BY a.assessment_date DESC,a.id DESC LIMIT 10`, filterArgs(filter)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []RecentResult
	for rows.Next() {
		var item RecentResult
		if err := rows.Scan(&item.EmployeeName, &item.EmployeeNumber, &item.PositionName, &item.DepartmentName, &item.StoreName, &item.TrainingMonth, &item.PeriodName, &item.AssessmentDate, &item.AssessorName, &item.AverageScore, &item.Rating); err != nil {
			return nil, fmt.Errorf("scan recent result: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) FilterOptions(ctx context.Context) ([]DepartmentOption, []RubricOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT d.id,d.name FROM departments d JOIN trainee_assessment_rubrics b ON b.department_id=d.id WHERE b.is_active=1 ORDER BY d.name`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var departments []DepartmentOption
	for rows.Next() {
		var x DepartmentOption
		if err := rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, nil, err
		}
		departments = append(departments, x)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	periodRows, err := r.db.QueryContext(ctx, `SELECT b.id,b.department_id,b.name,CASE WHEN b.month_to IS NULL THEN CONCAT('Bulan ',b.month_from,'+') WHEN b.month_from=b.month_to THEN CONCAT('Bulan ',b.month_from) ELSE CONCAT('Bulan ',b.month_from,'–',b.month_to) END FROM trainee_assessment_rubrics b WHERE b.is_active=1 ORDER BY b.department_id,b.month_from`)
	if err != nil {
		return nil, nil, err
	}
	defer periodRows.Close()
	var rubrics []RubricOption
	for periodRows.Next() {
		var x RubricOption
		if err := periodRows.Scan(&x.ID, &x.DepartmentID, &x.Name, &x.PeriodName); err != nil {
			return nil, nil, err
		}
		rubrics = append(rubrics, x)
	}
	return departments, rubrics, periodRows.Err()
}

package assessmentcriterion

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Criteria(ctx context.Context) ([]Criterion, error) {
	const query = `SELECT c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,c.department_id,dp.name,
		COUNT(sr.score),5*(SELECT COUNT(*) FROM trainee_assessment_rubrics rb WHERE rb.department_id=c.department_id AND rb.is_active=1),COUNT(DISTINCT d.assessment_id)
		FROM trainee_assessment_criteria c
		JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN departments dp ON dp.id=c.department_id
		LEFT JOIN trainee_assessment_score_rules sr ON sr.criterion_id=c.id
		LEFT JOIN trainee_assessment_details d ON d.criterion_id=c.id
		GROUP BY c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,c.department_id,dp.name
		ORDER BY dp.name,a.sort_order,c.sort_order,c.id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assessment criteria: %w", err)
	}
	defer rows.Close()
	var result []Criterion
	for rows.Next() {
		var item Criterion
		if err := rows.Scan(&item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.DepartmentID, &item.DepartmentName, &item.RuleCount, &item.ExpectedRuleCount, &item.AssessmentCount); err != nil {
			return nil, fmt.Errorf("scan assessment criterion: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Aspects(ctx context.Context) ([]AspectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM trainee_assessment_aspects ORDER BY sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list aspect options: %w", err)
	}
	defer rows.Close()
	var result []AspectOption
	for rows.Next() {
		var item AspectOption
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Criterion(ctx context.Context, id uint64) (Criterion, error) {
	var item Criterion
	err := r.db.QueryRowContext(ctx, `SELECT c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,c.department_id,dp.name,
		(SELECT COUNT(*) FROM trainee_assessment_score_rules sr WHERE sr.criterion_id=c.id),
		5*(SELECT COUNT(*) FROM trainee_assessment_rubrics rb WHERE rb.department_id=c.department_id AND rb.is_active=1),
		(SELECT COUNT(*) FROM trainee_assessment_details d WHERE d.criterion_id=c.id)
		FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN departments dp ON dp.id=c.department_id WHERE c.id=?`, id).
		Scan(&item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.DepartmentID, &item.DepartmentName, &item.RuleCount, &item.ExpectedRuleCount, &item.AssessmentCount)
	return item, err
}

func (r *Repository) AspectExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_assessment_aspects WHERE id=?)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) DepartmentExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM departments WHERE id=?)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) Departments(ctx context.Context) ([]AspectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM departments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AspectOption
	for rows.Next() {
		var x AspectOption
		if err := rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, err
		}
		result = append(result, x)
	}
	return result, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_criteria(department_id,aspect_id,code,description,weight,sort_order) VALUES(?,?,?,?,?,?)`, input.DepartmentID, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	_, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_criteria SET department_id=?,aspect_id=?,code=?,description=?,weight=?,sort_order=? WHERE id=?`, input.DepartmentID, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder, id)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM trainee_assessment_score_rules WHERE criterion_id=?`, id); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM trainee_assessment_criteria WHERE id=?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (r *Repository) Rubrics(ctx context.Context) ([]RubricOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,d.name,r.name,CASE WHEN r.month_to IS NULL THEN CONCAT('Bulan ',r.month_from,'+') WHEN r.month_from=r.month_to THEN CONCAT('Bulan ',r.month_from) ELSE CONCAT('Bulan ',r.month_from,'–',r.month_to) END FROM trainee_assessment_rubrics r JOIN departments d ON d.id=r.department_id WHERE r.is_active=1 ORDER BY d.name,r.month_from`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []RubricOption
	for rows.Next() {
		var x RubricOption
		if err := rows.Scan(&x.ID, &x.DepartmentName, &x.Name, &x.MonthLabel); err != nil {
			return nil, err
		}
		result = append(result, x)
	}
	return result, rows.Err()
}

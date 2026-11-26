package assessmentcriterion

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Criteria(ctx context.Context) ([]Criterion, error) {
	const query = `SELECT COALESCE(c.question_group_id,0),COALESCE((SELECT name FROM assessment_question_groups WHERE id=c.question_group_id),''),c.employment_status,c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,MIN(cp.position_id),GROUP_CONCAT(DISTINCT p.name ORDER BY p.name SEPARATOR ', '),GROUP_CONCAT(DISTINCT cp.position_id ORDER BY cp.position_id),
		COUNT(DISTINCT CONCAT(sr.rubric_id,':',sr.score)),5*(SELECT COUNT(*) FROM trainee_assessment_rubrics rb JOIN trainee_assessment_criterion_positions cx ON cx.position_id=rb.position_id WHERE cx.criterion_id=c.id AND rb.is_active=1),COUNT(DISTINCT d.assessment_id)
		FROM trainee_assessment_criteria c
		JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN trainee_assessment_criterion_positions cp ON cp.criterion_id=c.id JOIN positions p ON p.id=cp.position_id
		LEFT JOIN trainee_assessment_score_rules sr ON sr.criterion_id=c.id
		LEFT JOIN trainee_assessment_details d ON d.criterion_id=c.id
		GROUP BY c.question_group_id,c.employment_status,c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order
		ORDER BY MIN(p.name),a.sort_order,c.sort_order,c.id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list assessment criteria: %w", err)
	}
	defer rows.Close()
	var result []Criterion
	for rows.Next() {
		var item Criterion
		if err := rows.Scan(&item.QuestionGroupID, &item.QuestionGroupName, &item.EmploymentStatus, &item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.DepartmentID, &item.DepartmentName, &item.PositionIDs, &item.RuleCount, &item.ExpectedRuleCount, &item.AssessmentCount); err != nil {
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
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(c.question_group_id,0),COALESCE((SELECT name FROM assessment_question_groups WHERE id=c.question_group_id),''),c.employment_status,c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order,MIN(cp.position_id),GROUP_CONCAT(DISTINCT p.name ORDER BY p.name SEPARATOR ', '),GROUP_CONCAT(DISTINCT cp.position_id ORDER BY cp.position_id),
		(SELECT COUNT(*) FROM trainee_assessment_score_rules sr WHERE sr.criterion_id=c.id),
		5*(SELECT COUNT(*) FROM trainee_assessment_rubrics rb JOIN trainee_assessment_criterion_positions cx ON cx.position_id=rb.position_id WHERE cx.criterion_id=c.id AND rb.is_active=1),
		(SELECT COUNT(*) FROM trainee_assessment_details d WHERE d.criterion_id=c.id)
		FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN trainee_assessment_criterion_positions cp ON cp.criterion_id=c.id JOIN positions p ON p.id=cp.position_id WHERE c.id=? GROUP BY c.question_group_id,c.employment_status,c.id,c.aspect_id,a.code,a.name,c.code,c.description,c.weight,c.sort_order`, id).
		Scan(&item.QuestionGroupID, &item.QuestionGroupName, &item.EmploymentStatus, &item.ID, &item.AspectID, &item.AspectCode, &item.AspectName, &item.Code, &item.Description, &item.Weight, &item.SortOrder, &item.DepartmentID, &item.DepartmentName, &item.PositionIDs, &item.RuleCount, &item.ExpectedRuleCount, &item.AssessmentCount)
	return item, err
}

func (r *Repository) AspectExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_assessment_aspects WHERE id=?)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) DepartmentExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM positions WHERE id=?)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) Departments(ctx context.Context) ([]AspectOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name FROM positions ORDER BY name`)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO trainee_assessment_criteria(position_id,department_id,question_group_id,employment_status,aspect_id,code,description,weight,sort_order) SELECT p.id,p.department_id,NULLIF(?,0),?,?,?,?,?,? FROM positions p WHERE p.id=?`, input.QuestionGroupID, input.EmploymentStatus, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder, input.DepartmentID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	for _, positionID := range input.PositionIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO trainee_assessment_criterion_positions(criterion_id,position_id) VALUES(?,?)`, id, positionID); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE sr FROM trainee_assessment_score_rules sr JOIN trainee_assessment_rubrics r ON r.id=sr.rubric_id WHERE sr.criterion_id=? AND NOT EXISTS(SELECT 1 FROM trainee_assessment_criterion_positions cp WHERE cp.criterion_id=sr.criterion_id AND cp.position_id=r.position_id)`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `UPDATE trainee_assessment_criteria c JOIN positions p ON p.id=? SET c.position_id=p.id,c.department_id=p.department_id,c.question_group_id=NULLIF(?,0),c.employment_status=?,c.aspect_id=?,c.code=?,c.description=?,c.weight=?,c.sort_order=? WHERE c.id=?`, input.DepartmentID, input.QuestionGroupID, input.EmploymentStatus, input.AspectID, input.Code, input.Description, input.Weight, input.SortOrder, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM trainee_assessment_criterion_positions WHERE criterion_id=?`, id); err != nil {
		return err
	}
	for _, positionID := range input.PositionIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO trainee_assessment_criterion_positions(criterion_id,position_id) VALUES(?,?)`, id, positionID); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE sr FROM trainee_assessment_score_rules sr JOIN trainee_assessment_rubrics r ON r.id=sr.rubric_id WHERE sr.criterion_id=? AND NOT EXISTS(SELECT 1 FROM trainee_assessment_criterion_positions cp WHERE cp.criterion_id=sr.criterion_id AND cp.position_id=r.position_id)`, id); err != nil {
		return err
	}
	return tx.Commit()
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
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,p.name,r.name,CASE WHEN r.month_to IS NULL THEN CONCAT('Bulan ',r.month_from,'+') WHEN r.month_from=r.month_to THEN CONCAT('Bulan ',r.month_from) ELSE CONCAT('Bulan ',r.month_from,'–',r.month_to) END FROM trainee_assessment_rubrics r JOIN positions p ON p.id=r.position_id WHERE r.is_active=1 ORDER BY p.name,r.month_from`)
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

func (r *Repository) QuestionGroups(ctx context.Context) ([]AspectOption, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,name FROM assessment_question_groups ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AspectOption
	for rows.Next() {
		var x AspectOption
		if err = rows.Scan(&x.ID, &x.Name); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

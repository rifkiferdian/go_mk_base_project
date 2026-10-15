package assessmentrule

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const ruleSelect = `SELECT r.rubric_id,b.name,d.name,CASE WHEN b.month_to IS NULL THEN CONCAT('Bulan ',b.month_from,'+') ELSE CONCAT('Bulan ',b.month_from,'–',b.month_to) END,r.criterion_id,c.code,c.description,a.id,a.name,r.score,r.finding_min,r.finding_max,r.description FROM trainee_assessment_score_rules r JOIN trainee_assessment_criteria c ON c.id=r.criterion_id JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN trainee_assessment_rubrics b ON b.id=r.rubric_id JOIN departments d ON d.id=b.department_id`

func scanRule(scanner interface{ Scan(...any) error }) (Rule, error) {
	var x Rule
	var maximum sql.NullInt64
	err := scanner.Scan(&x.RubricID, &x.RubricName, &x.DepartmentName, &x.MonthLabel, &x.CriterionID, &x.CriterionCode, &x.CriterionDescription, &x.AspectID, &x.AspectName, &x.Score, &x.FindingMin, &maximum, &x.Description)
	x.HasFindingMax = maximum.Valid
	if maximum.Valid {
		x.FindingMax = uint16(maximum.Int64)
	}
	return x, err
}
func (r *Repository) Rules(ctx context.Context) ([]Rule, error) {
	rows, err := r.db.QueryContext(ctx, ruleSelect+` ORDER BY d.name,b.month_from,a.sort_order,c.sort_order,r.score DESC`)
	if err != nil {
		return nil, fmt.Errorf("list assessment score rules: %w", err)
	}
	defer rows.Close()
	var out []Rule
	for rows.Next() {
		x, e := scanRule(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Criteria(ctx context.Context) ([]CriterionOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT c.id,a.id,a.name,c.code,c.description,COUNT(r.score),5*(SELECT COUNT(*) FROM trainee_assessment_rubrics rb WHERE rb.department_id=c.department_id AND rb.is_active=1),c.department_id,d.name FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id JOIN departments d ON d.id=c.department_id LEFT JOIN trainee_assessment_score_rules r ON r.criterion_id=c.id GROUP BY c.id,a.id,a.name,c.code,c.description,c.department_id,d.name,a.sort_order,c.sort_order ORDER BY d.name,a.sort_order,c.sort_order,c.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CriterionOption
	for rows.Next() {
		var x CriterionOption
		if err := rows.Scan(&x.ID, &x.AspectID, &x.AspectName, &x.Code, &x.Description, &x.RuleCount, &x.ExpectedRuleCount, &x.DepartmentID, &x.DepartmentName); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Rubrics(ctx context.Context) ([]RubricOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,r.department_id,r.name,d.name,CASE WHEN r.month_to IS NULL THEN CONCAT('Bulan ',r.month_from,'+') ELSE CONCAT('Bulan ',r.month_from,'–',r.month_to) END FROM trainee_assessment_rubrics r JOIN departments d ON d.id=r.department_id WHERE r.is_active=1 ORDER BY d.name,r.month_from`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RubricOption
	for rows.Next() {
		var x RubricOption
		if err := rows.Scan(&x.ID, &x.DepartmentID, &x.Name, &x.DepartmentName, &x.MonthLabel); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Rule(ctx context.Context, key Key) (Rule, error) {
	return scanRule(r.db.QueryRowContext(ctx, ruleSelect+` WHERE r.rubric_id=? AND r.criterion_id=? AND r.score=?`, key.RubricID, key.CriterionID, key.Score))
}
func (r *Repository) CriterionMatchesRubric(ctx context.Context, criterionID, rubricID uint64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_assessment_criteria c JOIN trainee_assessment_rubrics b ON b.department_id=c.department_id WHERE c.id=? AND b.id=? AND b.is_active=1)`, criterionID, rubricID).Scan(&ok)
	return ok, err
}
func (r *Repository) RulesForCriterion(ctx context.Context, rubricID, criterionID uint64) ([]Rule, error) {
	rows, err := r.db.QueryContext(ctx, ruleSelect+` WHERE r.rubric_id=? AND r.criterion_id=?`, rubricID, criterionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Rule
	for rows.Next() {
		x, e := scanRule(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Create(ctx context.Context, x Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description) VALUES(?,?,?,?,?,?)`, x.RubricID, x.CriterionID, x.Score, x.FindingMin, x.FindingMax, x.Description)
	return err
}
func (r *Repository) Update(ctx context.Context, old Key, x Input) error {
	res, err := r.db.ExecContext(ctx, `UPDATE trainee_assessment_score_rules SET rubric_id=?,criterion_id=?,score=?,finding_min=?,finding_max=?,description=? WHERE rubric_id=? AND criterion_id=? AND score=?`, x.RubricID, x.CriterionID, x.Score, x.FindingMin, x.FindingMax, x.Description, old.RubricID, old.CriterionID, old.Score)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		if _, e := r.Rule(ctx, Key{x.RubricID, x.CriterionID, x.Score}); e == nil {
			return nil
		}
		return sql.ErrNoRows
	}
	return nil
}
func (r *Repository) Delete(ctx context.Context, key Key) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM trainee_assessment_score_rules WHERE rubric_id=? AND criterion_id=? AND score=?`, key.RubricID, key.CriterionID, key.Score)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

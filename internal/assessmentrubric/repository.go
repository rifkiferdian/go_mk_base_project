package assessmentrubric

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Rubrics(ctx context.Context) ([]Rubric, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT b.id,b.position_id,p.department_id,p.name,d.name,b.code,b.name,b.month_from,COALESCE(b.month_to,0),b.is_active,COUNT(DISTINCT sr.criterion_id),COUNT(DISTINCT a.id) FROM trainee_assessment_rubrics b JOIN positions p ON p.id=b.position_id JOIN departments d ON d.id=p.department_id LEFT JOIN trainee_assessment_score_rules sr ON sr.rubric_id=b.id LEFT JOIN trainee_assessments a ON a.rubric_id=b.id GROUP BY b.id,b.position_id,p.department_id,p.name,d.name,b.code,b.name,b.month_from,b.month_to,b.is_active ORDER BY p.name,b.month_from`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Rubric
	for rows.Next() {
		var x Rubric
		if e = rows.Scan(&x.ID, &x.PositionID, &x.DepartmentID, &x.PositionName, &x.DepartmentName, &x.Code, &x.Name, &x.MonthFrom, &x.MonthTo, &x.IsActive, &x.RuleCount, &x.AssessmentCount); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Rubric(ctx context.Context, id uint64) (Rubric, error) {
	var x Rubric
	e := r.db.QueryRowContext(ctx, `SELECT b.id,b.position_id,p.department_id,p.name,d.name,b.code,b.name,b.month_from,COALESCE(b.month_to,0),b.is_active,(SELECT COUNT(DISTINCT criterion_id) FROM trainee_assessment_score_rules WHERE rubric_id=b.id),(SELECT COUNT(*) FROM trainee_assessments WHERE rubric_id=b.id) FROM trainee_assessment_rubrics b JOIN positions p ON p.id=b.position_id JOIN departments d ON d.id=p.department_id WHERE b.id=?`, id).Scan(&x.ID, &x.PositionID, &x.DepartmentID, &x.PositionName, &x.DepartmentName, &x.Code, &x.Name, &x.MonthFrom, &x.MonthTo, &x.IsActive, &x.RuleCount, &x.AssessmentCount)
	return x, e
}
func (r *Repository) Positions(ctx context.Context) ([]Position, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,CONCAT(name,' — ',(SELECT name FROM departments WHERE id=positions.department_id)) FROM positions ORDER BY name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Position
	for rows.Next() {
		var x Position
		if e = rows.Scan(&x.ID, &x.Name); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Overlaps(ctx context.Context, id, positionID uint64, from, to uint16) (bool, error) {
	var n int
	e := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trainee_assessment_rubrics WHERE id<>? AND position_id=? AND month_from<=? AND COALESCE(month_to,65535)>=?`, id, positionID, to, from).Scan(&n)
	return n > 0, e
}
func (r *Repository) Create(ctx context.Context, x Input) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active) SELECT p.department_id,p.id,CONCAT('period_',?,'_',?),CONCAT(p.name,' — Bulan ',?,'–',?),?,?,? FROM positions p WHERE p.id=?`, x.MonthFrom, x.MonthTo, x.MonthFrom, x.MonthTo, x.MonthFrom, x.MonthTo, x.IsActive, x.PositionID)
	return e
}
func (r *Repository) Update(ctx context.Context, id uint64, x Input) error {
	res, e := r.db.ExecContext(ctx, `UPDATE trainee_assessment_rubrics b JOIN positions p ON p.id=? SET b.department_id=p.department_id,b.position_id=p.id,b.name=CONCAT(p.name,' — Bulan ',?,'–',?),b.month_from=?,b.month_to=?,b.is_active=? WHERE b.id=?`, x.PositionID, x.MonthFrom, x.MonthTo, x.MonthFrom, x.MonthTo, x.IsActive, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, e = r.Rubric(ctx, id)
	}
	return e
}
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM trainee_assessment_rubrics WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

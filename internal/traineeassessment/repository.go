package traineeassessment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Assessments(ctx context.Context, userID uint64, all bool) ([]Assessment, error) {
	q := `SELECT a.id,a.trainee_employee_id,a.employee_number,a.employee_name,COALESCE(a.trainee_position_name,''),a.department_name,a.store_name,a.training_month,DATE_FORMAT(a.assessment_date,'%Y-%m-%d'),COALESCE(a.notes,''),a.status,a.assessor_name,COALESCE(a.assessor_position_name,''),COALESCE(v.total_score,0),COALESCE(v.total_weight,0),COALESCE(v.average_score,0),COALESCE(v.rating,''),COALESCE(v.scored_count,0),COALESCE(v.criterion_count,0) FROM trainee_assessments a LEFT JOIN trainee_assessment_results v ON v.assessment_id=a.id`
	args := []any{}
	if !all {
		q += ` JOIN users au ON au.id=? JOIN employee_assignments aa ON aa.employee_id=au.employee_id JOIN employee_reporting_lines rl ON rl.supervisor_assignment_id=aa.id AND rl.employee_assignment_id=a.trainee_assignment_id WHERE rl.started_at<=a.assessment_date AND (rl.ended_at IS NULL OR rl.ended_at>=a.assessment_date)`
		args = append(args, userID)
	}
	q += ` ORDER BY a.assessment_date DESC,a.id DESC`
	rows, e := r.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Assessment
	for rows.Next() {
		var x Assessment
		if e = rows.Scan(&x.ID, &x.TraineeEmployeeID, &x.EmployeeNumber, &x.EmployeeName, &x.PositionName, &x.DepartmentName, &x.StoreName, &x.TrainingMonth, &x.AssessmentDate, &x.Notes, &x.Status, &x.AssessorName, &x.AssessorPosition, &x.TotalScore, &x.TotalWeight, &x.AverageScore, &x.Rating, &x.ScoredCount, &x.CriterionCount); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Trainees(ctx context.Context, userID uint64, all bool) ([]Trainee, error) {
	q := `SELECT e.id,a.id,a.store_id,e.employee_number,e.name,p.name,d.name,s.name FROM employees e JOIN employee_profiles ep ON ep.employee_id=e.id AND ep.employment_status='trainee' JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id JOIN departments d ON d.id=a.department_id JOIN stores s ON s.id=a.store_id`
	args := []any{}
	if !all {
		q += ` JOIN users au ON au.id=? JOIN employee_assignments sa ON sa.employee_id=au.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.employee_assignment_id=a.id AND rl.supervisor_assignment_id=sa.id AND rl.ended_at IS NULL`
		args = append(args, userID)
	}
	q += ` WHERE e.status='active' ORDER BY e.name`
	rows, e := r.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Trainee
	for rows.Next() {
		var x Trainee
		if e = rows.Scan(&x.EmployeeID, &x.AssignmentID, &x.StoreID, &x.EmployeeNumber, &x.Name, &x.PositionName, &x.DepartmentName, &x.StoreName); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Rubric(ctx context.Context) ([]Aspect, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT a.id,a.name,c.id,c.description,c.weight,r.score,r.finding_min,r.finding_max,r.description FROM trainee_assessment_aspects a JOIN trainee_assessment_criteria c ON c.aspect_id=a.id LEFT JOIN trainee_assessment_score_rules r ON r.criterion_id=c.id ORDER BY a.sort_order,c.sort_order,r.score DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Aspect
	var ai, ci = -1, -1
	for rows.Next() {
		var aid, cid uint64
		var an, desc, rd string
		var weight float64
		var score int
		var min int
		var max sql.NullInt64
		if e = rows.Scan(&aid, &an, &cid, &desc, &weight, &score, &min, &max, &rd); e != nil {
			return nil, e
		}
		if ai < 0 || out[ai].ID != aid {
			out = append(out, Aspect{ID: aid, Name: an})
			ai = len(out) - 1
			ci = -1
		}
		if ci < 0 || out[ai].Criteria[ci].ID != cid {
			out[ai].Criteria = append(out[ai].Criteria, Criterion{ID: cid, AspectID: aid, AspectName: an, Description: desc, Weight: weight})
			ci = len(out[ai].Criteria) - 1
		}
		label := fmt.Sprintf("%d+ temuan", min)
		if max.Valid {
			if int64(min) == max.Int64 {
				label = fmt.Sprintf("%d temuan", min)
			} else {
				label = fmt.Sprintf("%d–%d temuan", min, max.Int64)
			}
		}
		out[ai].Criteria[ci].Rules = append(out[ai].Criteria[ci].Rules, Rule{score, rd, label})
	}
	return out, rows.Err()
}
func (r *Repository) Assessment(ctx context.Context, id, userID uint64, all bool) (Assessment, error) {
	items, e := r.Assessments(ctx, userID, all)
	if e != nil {
		return Assessment{}, e
	}
	for _, x := range items {
		if x.ID == id {
			return x, nil
		}
	}
	return Assessment{}, sql.ErrNoRows
}
func (r *Repository) DetailScores(ctx context.Context, id uint64) (map[uint64]int, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT criterion_id,COALESCE(score,0) FROM trainee_assessment_details WHERE assessment_id=?`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := map[uint64]int{}
	for rows.Next() {
		var cid uint64
		var score int
		if e = rows.Scan(&cid, &score); e != nil {
			return nil, e
		}
		out[cid] = score
	}
	return out, rows.Err()
}
func (r *Repository) AssessmentDetails(ctx context.Context, id uint64) ([]AssessmentDetail, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT aspect_name,criterion_description,weight,COALESCE(score,0) FROM trainee_assessment_details WHERE assessment_id=? ORDER BY criterion_id`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []AssessmentDetail
	for rows.Next() {
		var item AssessmentDetail
		if e = rows.Scan(&item.AspectName, &item.CriterionDescription, &item.Weight, &item.Score); e != nil {
			return nil, e
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (r *Repository) Create(ctx context.Context, userID uint64, assessorName string, input Input, all bool) (uint64, error) {
	return r.save(ctx, 0, userID, assessorName, input, all)
}
func (r *Repository) Update(ctx context.Context, id, userID uint64, assessorName string, input Input, all bool) (uint64, error) {
	return r.save(ctx, id, userID, assessorName, input, all)
}
func (r *Repository) save(ctx context.Context, id, userID uint64, assessorName string, input Input, all bool) (uint64, error) {
	tx, e := r.db.BeginTx(ctx, nil)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback()
	var t Trainee
	trainees, e := r.Trainees(ctx, userID, all)
	if e != nil {
		return 0, e
	}
	for _, x := range trainees {
		if x.EmployeeID == input.TraineeEmployeeID {
			t = x
		}
	}
	if t.EmployeeID == 0 {
		return 0, sql.ErrNoRows
	}
	var assessorAssignment sql.NullInt64
	var assessorPosition sql.NullString
	_ = tx.QueryRowContext(ctx, `SELECT a.id,p.name FROM users u JOIN employee_assignments a ON a.employee_id=u.employee_id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id WHERE u.id=?`, userID).Scan(&assessorAssignment, &assessorPosition)
	signed := any(nil)
	if input.Status == "final" {
		signed = "CURRENT_TIMESTAMP(6)"
	}
	_ = signed
	if id == 0 {
		res, e := tx.ExecContext(ctx, `INSERT INTO trainee_assessments(assessor_user_id,store_id,employee_name,employee_number,department_name,store_name,training_month,assessment_date,notes,status,assessor_name,assessor_signed_at,trainee_employee_id,trainee_assignment_id,assessor_assignment_id,trainee_position_name,assessor_position_name) VALUES(?,?,?,?,?,?,?,?,?,?,?,IF(?='final',CURRENT_TIMESTAMP(6),NULL),?,?,?,?,?)`, userID, t.StoreID, t.Name, t.EmployeeNumber, t.DepartmentName, t.StoreName, input.TrainingMonth, input.AssessmentDate, nullText(input.Notes), input.Status, assessorName, input.Status, t.EmployeeID, t.AssignmentID, nullInt(assessorAssignment), t.PositionName, nullString(assessorPosition))
		if e != nil {
			return 0, e
		}
		v, _ := res.LastInsertId()
		id = uint64(v)
	} else {
		res, e := tx.ExecContext(ctx, `UPDATE trainee_assessments SET store_id=?,employee_name=?,employee_number=?,department_name=?,store_name=?,training_month=?,assessment_date=?,notes=?,status=?,assessor_name=?,assessor_signed_at=IF(?='final',COALESCE(assessor_signed_at,CURRENT_TIMESTAMP(6)),NULL),trainee_employee_id=?,trainee_assignment_id=?,assessor_assignment_id=?,trainee_position_name=?,assessor_position_name=? WHERE id=?`, t.StoreID, t.Name, t.EmployeeNumber, t.DepartmentName, t.StoreName, input.TrainingMonth, input.AssessmentDate, nullText(input.Notes), input.Status, assessorName, input.Status, t.EmployeeID, t.AssignmentID, nullInt(assessorAssignment), t.PositionName, nullString(assessorPosition), id)
		if e != nil {
			return 0, e
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			var exists int
			if tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM trainee_assessments WHERE id=?`, id).Scan(&exists) != nil || exists == 0 {
				return 0, sql.ErrNoRows
			}
		}
		if _, e = tx.ExecContext(ctx, `DELETE FROM trainee_assessment_details WHERE assessment_id=?`, id); e != nil {
			return 0, e
		}
	}
	for _, d := range input.Details {
		var score any
		if d.Score > 0 {
			score = d.Score
		}
		_, e = tx.ExecContext(ctx, `INSERT INTO trainee_assessment_details(assessment_id,criterion_id,aspect_name,criterion_description,weight,score) SELECT ?,c.id,a.name,c.description,c.weight,? FROM trainee_assessment_criteria c JOIN trainee_assessment_aspects a ON a.id=c.aspect_id WHERE c.id=?`, id, score, d.CriterionID)
		if e != nil {
			return 0, e
		}
	}
	if e = tx.Commit(); e != nil {
		return 0, e
	}
	return id, nil
}
func (r *Repository) Delete(ctx context.Context, id, userID uint64, all bool) error {
	if _, e := r.Assessment(ctx, id, userID, all); e != nil {
		return e
	}
	res, e := r.db.ExecContext(ctx, `DELETE FROM trainee_assessments WHERE id=? AND status='draft'`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func nullText(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}
func nullInt(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}
func nullString(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

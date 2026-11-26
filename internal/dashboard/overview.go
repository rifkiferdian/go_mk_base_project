package dashboard

import (
	"context"
	"database/sql"
	"fmt"
)

type Overview struct {
	Employees, Trainee, HLPP, Covered, Missing, Groups, Criteria, Draft, Final int
	Coverage                                                                   float64
}

func loadOverview(ctx context.Context, db *sql.DB, uid uint64, admin bool) (Overview, error) {
	var o Overview
	scope := ""
	args := []any{}
	if !admin {
		scope = " AND EXISTS(SELECT 1 FROM users u JOIN employee_assignments sa ON sa.employee_id=u.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.supervisor_assignment_id=sa.id AND rl.employee_assignment_id=a.id AND rl.ended_at IS NULL WHERE u.id=?)"
		args = append(args, uid)
	}
	coverage := `EXISTS(SELECT 1 FROM trainee_assessment_criterion_positions cp JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id JOIN trainee_assessment_score_rules sr ON sr.criterion_id=c.id JOIN trainee_assessment_rubrics r ON r.id=sr.rubric_id AND r.position_id=cp.position_id AND r.is_active=1 WHERE cp.position_id=a.position_id AND c.employment_status=ep.employment_status)`
	q := fmt.Sprintf(`SELECT COUNT(*),COALESCE(SUM(ep.employment_status='trainee'),0),COALESCE(SUM(ep.employment_status='post_retirement_daily_worker'),0),COALESCE(SUM(%s),0) FROM employees e JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id LEFT JOIN employee_profiles ep ON ep.employee_id=e.id WHERE e.status='active' AND p.code NOT LIKE 'spv%%'`, coverage) + scope
	if e := db.QueryRowContext(ctx, q, args...).Scan(&o.Employees, &o.Trainee, &o.HLPP, &o.Covered); e != nil {
		return o, e
	}
	o.Missing = o.Employees - o.Covered
	if o.Employees > 0 {
		o.Coverage = float64(o.Covered) * 100 / float64(o.Employees)
	}
	if admin {
		if e := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assessment_question_groups").Scan(&o.Groups); e != nil {
			return o, e
		}
		if e := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM trainee_assessment_criteria").Scan(&o.Criteria); e != nil {
			return o, e
		}
	}
	return o, nil
}

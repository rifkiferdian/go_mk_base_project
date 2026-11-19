package team

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Members(ctx context.Context, userID uint64, all bool) ([]Member, error) {
	q := `SELECT e.id,e.employee_number,e.name,p.name,d.name,s.name,DATE_FORMAT(a.started_at,'%d-%m-%Y'),ep.employment_status FROM employees e JOIN employee_profiles ep ON ep.employee_id=e.id JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id JOIN departments d ON d.id=a.department_id JOIN stores s ON s.id=a.store_id`
	args := []any{}
	if !all {
		q += ` JOIN users u ON u.id=? JOIN employee_assignments sa ON sa.employee_id=u.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.employee_assignment_id=a.id AND rl.supervisor_assignment_id=sa.id AND rl.ended_at IS NULL`
		args = append(args, userID)
	}
	q += ` WHERE e.status='active' ORDER BY e.name`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var x Member
		if err = rows.Scan(&x.ID, &x.EmployeeNumber, &x.Name, &x.Position, &x.Department, &x.Store, &x.StartedAt, &x.Employment); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

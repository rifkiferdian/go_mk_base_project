package placementrequest

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Requests(ctx context.Context, userID uint64, all bool) ([]Request, error) {
	q := `SELECT q.id,q.employee_id,e.employee_number,e.name,cp.name,cs.name,np.name,ns.name,DATE_FORMAT(q.effective_date,'%d-%m-%Y'),q.reason,q.status,ru.name,COALESCE(q.decision_notes,''),COALESCE(du.name,''),DATE_FORMAT(q.created_at,'%d-%m-%Y %H:%i') FROM employee_placement_change_requests q JOIN employees e ON e.id=q.employee_id JOIN employee_assignments ca ON ca.id=q.current_assignment_id JOIN positions cp ON cp.id=ca.position_id JOIN stores cs ON cs.id=ca.store_id JOIN positions np ON np.id=q.requested_position_id JOIN stores ns ON ns.id=q.requested_store_id JOIN users ru ON ru.id=q.requested_by_user_id LEFT JOIN users du ON du.id=q.decided_by_user_id`
	args := []any{}
	if !all {
		q += ` WHERE q.requested_by_user_id=?`
		args = append(args, userID)
	}
	q += ` ORDER BY q.status='pending' DESC,q.created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Request
	for rows.Next() {
		var x Request
		if err = rows.Scan(&x.ID, &x.EmployeeID, &x.EmployeeNumber, &x.EmployeeName, &x.CurrentPosition, &x.CurrentStore, &x.RequestedPosition, &x.RequestedStore, &x.EffectiveDate, &x.Reason, &x.Status, &x.RequesterName, &x.DecisionNotes, &x.DeciderName, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Employees(ctx context.Context, userID uint64, all bool) ([]Option, error) {
	q := `SELECT e.id,CONCAT(e.employee_number,' — ',e.name,' (',p.name,')'),a.department_id FROM employees e JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL JOIN positions p ON p.id=a.position_id`
	args := []any{}
	if !all {
		q += ` JOIN users u ON u.id=? JOIN employee_assignments sa ON sa.employee_id=u.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.employee_assignment_id=a.id AND rl.supervisor_assignment_id=sa.id AND rl.ended_at IS NULL`
		args = append(args, userID)
	}
	q += ` WHERE e.status='active' ORDER BY e.name`
	return r.options(ctx, q, args...)
}
func (r *Repository) Positions(ctx context.Context) ([]Option, error) {
	return r.options(ctx, `SELECT p.id,CONCAT(p.name,' — ',d.name),p.department_id FROM positions p JOIN departments d ON d.id=p.department_id ORDER BY p.name`)
}
func (r *Repository) Stores(ctx context.Context) ([]Option, error) {
	return r.options(ctx, `SELECT id,name,0 FROM stores WHERE status='active' AND deleted_at IS NULL ORDER BY name`)
}
func (r *Repository) options(ctx context.Context, q string, args ...any) ([]Option, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Option
	for rows.Next() {
		var x Option
		if err = rows.Scan(&x.ID, &x.Name, &x.DepartmentID); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Create(ctx context.Context, userID uint64, x Input, all bool) error {
	var assignmentID uint64
	q := `SELECT a.id FROM employee_assignments a WHERE a.employee_id=? AND a.ended_at IS NULL`
	args := []any{x.EmployeeID}
	if !all {
		q += ` AND EXISTS(SELECT 1 FROM users u JOIN employee_assignments sa ON sa.employee_id=u.employee_id AND sa.ended_at IS NULL JOIN employee_reporting_lines rl ON rl.supervisor_assignment_id=sa.id AND rl.employee_assignment_id=a.id AND rl.ended_at IS NULL WHERE u.id=?)`
		args = append(args, userID)
	}
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&assignmentID); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO employee_placement_change_requests(employee_id,current_assignment_id,requested_position_id,requested_store_id,effective_date,reason,requested_by_user_id) VALUES(?,?,?,?,?,?,?)`, x.EmployeeID, assignmentID, x.PositionID, x.StoreID, x.EffectiveDate, x.Reason, userID)
	return err
}
func (r *Repository) Decide(ctx context.Context, id, userID uint64, approve bool, notes string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var employeeID, oldAssignment, positionID, storeID, departmentID uint64
	var effective time.Time
	var status string
	err = tx.QueryRowContext(ctx, `SELECT q.employee_id,q.current_assignment_id,q.requested_position_id,q.requested_store_id,q.effective_date,q.status,p.department_id FROM employee_placement_change_requests q JOIN positions p ON p.id=q.requested_position_id WHERE q.id=? FOR UPDATE`, id).Scan(&employeeID, &oldAssignment, &positionID, &storeID, &effective, &status, &departmentID)
	if err != nil {
		return err
	}
	if status != "pending" {
		return errors.New("request already decided")
	}
	if !approve {
		_, err = tx.ExecContext(ctx, `UPDATE employee_placement_change_requests SET status='rejected',decided_by_user_id=?,decision_notes=NULLIF(?,''),decided_at=CURRENT_TIMESTAMP(6) WHERE id=?`, userID, notes, id)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	today := time.Now().Truncate(24 * time.Hour)
	if effective.After(today) {
		return ErrFuture
	}
	var started time.Time
	if err = tx.QueryRowContext(ctx, `SELECT started_at FROM employee_assignments WHERE id=? AND employee_id=? AND ended_at IS NULL FOR UPDATE`, oldAssignment, employeeID).Scan(&started); err != nil {
		return err
	}
	end := effective.AddDate(0, 0, -1)
	if end.Before(started) {
		return ErrInvalid
	}
	if _, err = tx.ExecContext(ctx, `UPDATE employee_reporting_lines SET ended_at=? WHERE employee_assignment_id=? AND ended_at IS NULL`, end, oldAssignment); err != nil {
		return err
	}
	var supervisorIDs []uint64
	rows, e := tx.QueryContext(ctx, `SELECT supervisor_assignment_id FROM employee_reporting_lines WHERE employee_assignment_id=? AND ended_at=?`, oldAssignment, end)
	if e != nil {
		return e
	}
	for rows.Next() {
		var v uint64
		if e = rows.Scan(&v); e != nil {
			rows.Close()
			return e
		}
		supervisorIDs = append(supervisorIDs, v)
	}
	rows.Close()
	if _, err = tx.ExecContext(ctx, `UPDATE employee_assignments SET ended_at=? WHERE id=?`, end, oldAssignment); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO employee_assignments(employee_id,position_id,department_id,store_id,started_at) VALUES(?,?,?,?,?)`, employeeID, positionID, departmentID, storeID, effective)
	if err != nil {
		return err
	}
	newID, _ := res.LastInsertId()
	for _, sid := range supervisorIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO employee_reporting_lines(employee_assignment_id,supervisor_assignment_id,started_at) VALUES(?,?,?)`, newID, sid, effective); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE employee_placement_change_requests SET status='approved',decided_by_user_id=?,decision_notes=NULLIF(?,''),decided_at=CURRENT_TIMESTAMP(6) WHERE id=?`, userID, notes, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

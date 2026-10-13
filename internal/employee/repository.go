package employee

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Employees(ctx context.Context) ([]Employee, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id,e.employee_number,e.name,COALESCE(e.email,''),COALESCE(e.phone,''),e.status,
	 COALESCE(ep.employment_status,''),COALESCE(p.name,''),COALESCE(d.name,''),COALESCE(s.name,''),ep.employee_id IS NOT NULL,a.id IS NOT NULL,u.id IS NOT NULL,
	 (SELECT COUNT(*) FROM employee_assignments ah WHERE ah.employee_id=e.id),
	 (SELECT COUNT(*) FROM employee_reporting_lines rl JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id WHERE sa.employee_id=e.id AND rl.ended_at IS NULL)
	 FROM employees e LEFT JOIN employee_profiles ep ON ep.employee_id=e.id LEFT JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL
	 LEFT JOIN positions p ON p.id=a.position_id LEFT JOIN departments d ON d.id=a.department_id LEFT JOIN stores s ON s.id=a.store_id LEFT JOIN users u ON u.employee_id=e.id ORDER BY e.name`)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()
	var result []Employee
	for rows.Next() {
		var x Employee
		if err := rows.Scan(&x.ID, &x.EmployeeNumber, &x.Name, &x.Email, &x.Phone, &x.Status, &x.EmploymentStatus, &x.PositionName, &x.DepartmentName, &x.StoreName, &x.HasProfile, &x.HasCurrentAssignment, &x.HasUser, &x.AssignmentCount, &x.SubordinateCount); err != nil {
			return nil, err
		}
		result = append(result, x)
	}
	return result, rows.Err()
}
func (r *Repository) Employee(ctx context.Context, id uint64) (Employee, error) {
	var x Employee
	err := r.db.QueryRowContext(ctx, `SELECT e.id,e.employee_number,e.name,COALESCE(e.email,''),COALESCE(e.phone,''),e.status,COALESCE(ep.employment_status,''),COALESCE(p.name,''),COALESCE(d.name,''),COALESCE(s.name,''),ep.employee_id IS NOT NULL,a.id IS NOT NULL,u.id IS NOT NULL,(SELECT COUNT(*) FROM employee_assignments ah WHERE ah.employee_id=e.id),(SELECT COUNT(*) FROM employee_reporting_lines rl JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id WHERE sa.employee_id=e.id AND rl.ended_at IS NULL) FROM employees e LEFT JOIN employee_profiles ep ON ep.employee_id=e.id LEFT JOIN employee_assignments a ON a.employee_id=e.id AND a.ended_at IS NULL LEFT JOIN positions p ON p.id=a.position_id LEFT JOIN departments d ON d.id=a.department_id LEFT JOIN stores s ON s.id=a.store_id LEFT JOIN users u ON u.employee_id=e.id WHERE e.id=?`, id).Scan(&x.ID, &x.EmployeeNumber, &x.Name, &x.Email, &x.Phone, &x.Status, &x.EmploymentStatus, &x.PositionName, &x.DepartmentName, &x.StoreName, &x.HasProfile, &x.HasCurrentAssignment, &x.HasUser, &x.AssignmentCount, &x.SubordinateCount)
	return x, err
}
func (r *Repository) CreateEmployee(ctx context.Context, x EmployeeInput) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO employees(employee_number,name,email,phone,status) VALUES(?,?,?,?,?)`, x.EmployeeNumber, x.Name, nullableString(x.Email), nullableString(x.Phone), x.Status)
	return e
}
func (r *Repository) UpdateEmployee(ctx context.Context, id uint64, x EmployeeInput) error {
	_, e := r.db.ExecContext(ctx, `UPDATE employees SET employee_number=?,name=?,email=?,phone=?,status=? WHERE id=?`, x.EmployeeNumber, x.Name, nullableString(x.Email), nullableString(x.Phone), x.Status, id)
	return e
}
func (r *Repository) DeleteEmployee(ctx context.Context, id uint64) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM employees WHERE id=?`, id)
	return affected(res, e)
}

func (r *Repository) Profiles(ctx context.Context) ([]Profile, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT ep.employee_id,e.employee_number,e.name,COALESCE(ep.birth_place,''),COALESCE(DATE_FORMAT(ep.birth_date,'%Y-%m-%d'),''),COALESCE(ep.gender,''),COALESCE(ep.address,''),COALESCE(DATE_FORMAT(ep.joined_date,'%Y-%m-%d'),''),ep.employment_status,COALESCE(ep.photo_path,''),COALESCE(ep.emergency_contact_name,''),COALESCE(ep.emergency_contact_phone,''),COALESCE(ep.emergency_contact_relationship,'') FROM employee_profiles ep JOIN employees e ON e.id=ep.employee_id ORDER BY e.name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		var x Profile
		if e = rows.Scan(&x.EmployeeID, &x.EmployeeNumber, &x.EmployeeName, &x.BirthPlace, &x.BirthDate, &x.Gender, &x.Address, &x.JoinedDate, &x.EmploymentStatus, &x.PhotoPath, &x.EmergencyContactName, &x.EmergencyContactPhone, &x.EmergencyContactRelationship); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Profile(ctx context.Context, id uint64) (Profile, error) {
	var x Profile
	e := r.db.QueryRowContext(ctx, `SELECT ep.employee_id,e.employee_number,e.name,COALESCE(ep.birth_place,''),COALESCE(DATE_FORMAT(ep.birth_date,'%Y-%m-%d'),''),COALESCE(ep.gender,''),COALESCE(ep.address,''),COALESCE(DATE_FORMAT(ep.joined_date,'%Y-%m-%d'),''),ep.employment_status,COALESCE(ep.photo_path,''),COALESCE(ep.emergency_contact_name,''),COALESCE(ep.emergency_contact_phone,''),COALESCE(ep.emergency_contact_relationship,'') FROM employee_profiles ep JOIN employees e ON e.id=ep.employee_id WHERE ep.employee_id=?`, id).Scan(&x.EmployeeID, &x.EmployeeNumber, &x.EmployeeName, &x.BirthPlace, &x.BirthDate, &x.Gender, &x.Address, &x.JoinedDate, &x.EmploymentStatus, &x.PhotoPath, &x.EmergencyContactName, &x.EmergencyContactPhone, &x.EmergencyContactRelationship)
	return x, e
}
func (r *Repository) CreateProfile(ctx context.Context, x ProfileInput) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO employee_profiles(employee_id,birth_place,birth_date,gender,address,joined_date,employment_status,photo_path,emergency_contact_name,emergency_contact_phone,emergency_contact_relationship) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, x.EmployeeID, nullableString(x.BirthPlace), nullableString(x.BirthDate), nullableString(x.Gender), nullableString(x.Address), nullableString(x.JoinedDate), x.EmploymentStatus, nullableString(x.PhotoPath), nullableString(x.EmergencyContactName), nullableString(x.EmergencyContactPhone), nullableString(x.EmergencyContactRelationship))
	return e
}
func (r *Repository) UpdateProfile(ctx context.Context, id uint64, x ProfileInput) error {
	_, e := r.db.ExecContext(ctx, `UPDATE employee_profiles SET birth_place=?,birth_date=?,gender=?,address=?,joined_date=?,employment_status=?,photo_path=?,emergency_contact_name=?,emergency_contact_phone=?,emergency_contact_relationship=? WHERE employee_id=?`, nullableString(x.BirthPlace), nullableString(x.BirthDate), nullableString(x.Gender), nullableString(x.Address), nullableString(x.JoinedDate), x.EmploymentStatus, nullableString(x.PhotoPath), nullableString(x.EmergencyContactName), nullableString(x.EmergencyContactPhone), nullableString(x.EmergencyContactRelationship), id)
	return e
}
func (r *Repository) DeleteProfile(ctx context.Context, id uint64) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM employee_profiles WHERE employee_id=?`, id)
	return affected(res, e)
}

func (r *Repository) Assignments(ctx context.Context) ([]Assignment, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT a.id,a.employee_id,a.position_id,a.department_id,a.store_id,e.employee_number,e.name,p.name,l.name,d.name,s.name,DATE_FORMAT(a.started_at,'%Y-%m-%d'),COALESCE(DATE_FORMAT(a.ended_at,'%Y-%m-%d'),''),a.ended_at IS NULL FROM employee_assignments a JOIN employees e ON e.id=a.employee_id JOIN positions p ON p.id=a.position_id JOIN job_levels l ON l.id=p.job_level_id JOIN departments d ON d.id=a.department_id JOIN stores s ON s.id=a.store_id ORDER BY a.started_at DESC,e.name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Assignment
	for rows.Next() {
		var x Assignment
		if e = rows.Scan(&x.ID, &x.EmployeeID, &x.PositionID, &x.DepartmentID, &x.StoreID, &x.EmployeeNumber, &x.EmployeeName, &x.PositionName, &x.JobLevelName, &x.DepartmentName, &x.StoreName, &x.StartedAt, &x.EndedAt, &x.Current); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) Assignment(ctx context.Context, id uint64) (Assignment, error) {
	var x Assignment
	e := r.db.QueryRowContext(ctx, `SELECT a.id,a.employee_id,a.position_id,a.department_id,a.store_id,e.employee_number,e.name,p.name,l.name,d.name,s.name,DATE_FORMAT(a.started_at,'%Y-%m-%d'),COALESCE(DATE_FORMAT(a.ended_at,'%Y-%m-%d'),''),a.ended_at IS NULL FROM employee_assignments a JOIN employees e ON e.id=a.employee_id JOIN positions p ON p.id=a.position_id JOIN job_levels l ON l.id=p.job_level_id JOIN departments d ON d.id=a.department_id JOIN stores s ON s.id=a.store_id WHERE a.id=?`, id).Scan(&x.ID, &x.EmployeeID, &x.PositionID, &x.DepartmentID, &x.StoreID, &x.EmployeeNumber, &x.EmployeeName, &x.PositionName, &x.JobLevelName, &x.DepartmentName, &x.StoreName, &x.StartedAt, &x.EndedAt, &x.Current)
	return x, e
}
func (r *Repository) CreateAssignment(ctx context.Context, x AssignmentInput) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO employee_assignments(employee_id,position_id,department_id,store_id,started_at,ended_at) VALUES(?,?,?,?,?,?)`, x.EmployeeID, x.PositionID, x.DepartmentID, x.StoreID, x.StartedAt, nullableString(x.EndedAt))
	return e
}
func (r *Repository) UpdateAssignment(ctx context.Context, id uint64, x AssignmentInput) error {
	_, e := r.db.ExecContext(ctx, `UPDATE employee_assignments SET employee_id=?,position_id=?,department_id=?,store_id=?,started_at=?,ended_at=? WHERE id=?`, x.EmployeeID, x.PositionID, x.DepartmentID, x.StoreID, x.StartedAt, nullableString(x.EndedAt), id)
	return e
}
func (r *Repository) DeleteAssignment(ctx context.Context, id uint64) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM employee_assignments WHERE id=?`, id)
	return affected(res, e)
}

func (r *Repository) ReportingLines(ctx context.Context) ([]ReportingLine, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT rl.id,rl.employee_assignment_id,rl.supervisor_assignment_id,e.employee_number,e.name,p.name,se.employee_number,se.name,sp.name,DATE_FORMAT(rl.started_at,'%Y-%m-%d'),COALESCE(DATE_FORMAT(rl.ended_at,'%Y-%m-%d'),''),rl.ended_at IS NULL FROM employee_reporting_lines rl JOIN employee_assignments a ON a.id=rl.employee_assignment_id JOIN employees e ON e.id=a.employee_id JOIN positions p ON p.id=a.position_id JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id JOIN employees se ON se.id=sa.employee_id JOIN positions sp ON sp.id=sa.position_id ORDER BY rl.started_at DESC,e.name`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []ReportingLine
	for rows.Next() {
		var x ReportingLine
		if e = rows.Scan(&x.ID, &x.EmployeeAssignmentID, &x.SupervisorAssignmentID, &x.EmployeeNumber, &x.EmployeeName, &x.EmployeePosition, &x.SupervisorNumber, &x.SupervisorName, &x.SupervisorPosition, &x.StartedAt, &x.EndedAt, &x.Current); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r *Repository) ReportingLine(ctx context.Context, id uint64) (ReportingLine, error) {
	var x ReportingLine
	e := r.db.QueryRowContext(ctx, `SELECT rl.id,rl.employee_assignment_id,rl.supervisor_assignment_id,e.employee_number,e.name,p.name,se.employee_number,se.name,sp.name,DATE_FORMAT(rl.started_at,'%Y-%m-%d'),COALESCE(DATE_FORMAT(rl.ended_at,'%Y-%m-%d'),''),rl.ended_at IS NULL FROM employee_reporting_lines rl JOIN employee_assignments a ON a.id=rl.employee_assignment_id JOIN employees e ON e.id=a.employee_id JOIN positions p ON p.id=a.position_id JOIN employee_assignments sa ON sa.id=rl.supervisor_assignment_id JOIN employees se ON se.id=sa.employee_id JOIN positions sp ON sp.id=sa.position_id WHERE rl.id=?`, id).Scan(&x.ID, &x.EmployeeAssignmentID, &x.SupervisorAssignmentID, &x.EmployeeNumber, &x.EmployeeName, &x.EmployeePosition, &x.SupervisorNumber, &x.SupervisorName, &x.SupervisorPosition, &x.StartedAt, &x.EndedAt, &x.Current)
	return x, e
}
func (r *Repository) CreateReportingLine(ctx context.Context, x ReportingLineInput) error {
	_, e := r.db.ExecContext(ctx, `INSERT INTO employee_reporting_lines(employee_assignment_id,supervisor_assignment_id,started_at,ended_at) VALUES(?,?,?,?)`, x.EmployeeAssignmentID, x.SupervisorAssignmentID, x.StartedAt, nullableString(x.EndedAt))
	return e
}
func (r *Repository) UpdateReportingLine(ctx context.Context, id uint64, x ReportingLineInput) error {
	_, e := r.db.ExecContext(ctx, `UPDATE employee_reporting_lines SET employee_assignment_id=?,supervisor_assignment_id=?,started_at=?,ended_at=? WHERE id=?`, x.EmployeeAssignmentID, x.SupervisorAssignmentID, x.StartedAt, nullableString(x.EndedAt), id)
	return e
}
func (r *Repository) DeleteReportingLine(ctx context.Context, id uint64) error {
	res, e := r.db.ExecContext(ctx, `DELETE FROM employee_reporting_lines WHERE id=?`, id)
	return affected(res, e)
}

func (r *Repository) Options(ctx context.Context, query string) ([]Option, error) {
	rows, e := r.db.QueryContext(ctx, query)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Option
	for rows.Next() {
		var x Option
		if e = rows.Scan(&x.ID, &x.Label); e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func affected(res sql.Result, e error) error {
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return e
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

package employee

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid employee data")
	ErrDuplicate = errors.New("duplicate employee data")
	ErrInUse     = errors.New("employee data in use")
)

type Service struct{ repository *Repository }

func NewService(r *Repository) *Service                            { return &Service{repository: r} }
func (s *Service) Employees(c context.Context) ([]Employee, error) { return s.repository.Employees(c) }
func (s *Service) Profiles(c context.Context) ([]Profile, error)   { return s.repository.Profiles(c) }
func (s *Service) Assignments(c context.Context) ([]Assignment, error) {
	return s.repository.Assignments(c)
}
func (s *Service) ReportingLines(c context.Context) ([]ReportingLine, error) {
	return s.repository.ReportingLines(c)
}
func (s *Service) EmployeeOptions(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT id,CONCAT(employee_number,' — ',name) FROM employees ORDER BY name`)
}
func (s *Service) EmployeesWithoutProfile(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT e.id,CONCAT(e.employee_number,' — ',e.name) FROM employees e LEFT JOIN employee_profiles p ON p.employee_id=e.id WHERE p.employee_id IS NULL ORDER BY e.name`)
}
func (s *Service) PositionOptions(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT p.id,CONCAT(l.name,' — ',p.name) FROM positions p JOIN job_levels l ON l.id=p.job_level_id ORDER BY l.rank_order,p.name`)
}
func (s *Service) DepartmentOptions(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT id,name FROM departments ORDER BY name`)
}
func (s *Service) StoreOptions(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT id,CONCAT(code,' — ',name) FROM stores WHERE deleted_at IS NULL AND status='active' ORDER BY name`)
}
func (s *Service) AssignmentOptions(c context.Context) ([]Option, error) {
	return s.repository.Options(c, `SELECT a.id,CONCAT(e.employee_number,' — ',e.name,' (',p.name,')') FROM employee_assignments a JOIN employees e ON e.id=a.employee_id JOIN positions p ON p.id=a.position_id ORDER BY a.ended_at IS NULL DESC,e.name,a.started_at DESC`)
}

func (s *Service) CreateEmployee(c context.Context, x EmployeeInput) error {
	x = cleanEmployee(x)
	if invalidEmployee(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.CreateEmployee(c, x))
}
func (s *Service) UpdateEmployee(c context.Context, id uint64, x EmployeeInput) error {
	if _, e := s.repository.Employee(c, id); e != nil {
		return e
	}
	x = cleanEmployee(x)
	if invalidEmployee(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.UpdateEmployee(c, id, x))
}
func (s *Service) DeleteEmployee(c context.Context, id uint64) error {
	x, e := s.repository.Employee(c, id)
	if e != nil {
		return e
	}
	if x.AssignmentCount > 0 || x.HasProfile || x.HasUser {
		return ErrInUse
	}
	return dbErr(s.repository.DeleteEmployee(c, id))
}
func cleanEmployee(x EmployeeInput) EmployeeInput {
	x.EmployeeNumber = strings.ToUpper(strings.TrimSpace(x.EmployeeNumber))
	x.Name = strings.TrimSpace(x.Name)
	x.Email = strings.ToLower(strings.TrimSpace(x.Email))
	x.Phone = strings.TrimSpace(x.Phone)
	x.Status = strings.TrimSpace(x.Status)
	return x
}
func invalidEmployee(x EmployeeInput) bool {
	if x.EmployeeNumber == "" || x.Name == "" || utf8.RuneCountInString(x.EmployeeNumber) > 50 || utf8.RuneCountInString(x.Name) > 150 || (x.Status != "active" && x.Status != "inactive") {
		return true
	}
	if x.Email != "" {
		_, e := mail.ParseAddress(x.Email)
		return e != nil || utf8.RuneCountInString(x.Email) > 191
	}
	return utf8.RuneCountInString(x.Phone) > 30
}

func (s *Service) CreateProfile(c context.Context, x ProfileInput) error {
	x = cleanProfile(x)
	if invalidProfile(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.CreateProfile(c, x))
}
func (s *Service) UpdateProfile(c context.Context, id uint64, x ProfileInput) error {
	if _, e := s.repository.Profile(c, id); e != nil {
		return e
	}
	x.EmployeeID = id
	x = cleanProfile(x)
	if invalidProfile(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.UpdateProfile(c, id, x))
}
func (s *Service) DeleteProfile(c context.Context, id uint64) error {
	return dbErr(s.repository.DeleteProfile(c, id))
}
func cleanProfile(x ProfileInput) ProfileInput {
	x.BirthPlace = strings.TrimSpace(x.BirthPlace)
	x.Gender = strings.TrimSpace(x.Gender)
	x.Address = strings.TrimSpace(x.Address)
	x.EmploymentStatus = strings.TrimSpace(x.EmploymentStatus)
	x.PhotoPath = strings.TrimSpace(x.PhotoPath)
	x.EmergencyContactName = strings.TrimSpace(x.EmergencyContactName)
	x.EmergencyContactPhone = strings.TrimSpace(x.EmergencyContactPhone)
	x.EmergencyContactRelationship = strings.TrimSpace(x.EmergencyContactRelationship)
	return x
}
func invalidProfile(x ProfileInput) bool {
	if x.EmployeeID == 0 || !oneOf(x.EmploymentStatus, "trainee", "probation", "contract", "permanent") || (!oneOf(x.Gender, "", "male", "female")) {
		return true
	}
	return !validDate(x.BirthDate, true) || !validDate(x.JoinedDate, true)
}

func (s *Service) CreateAssignment(c context.Context, x AssignmentInput) error {
	if invalidAssignment(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.CreateAssignment(c, x))
}
func (s *Service) UpdateAssignment(c context.Context, id uint64, x AssignmentInput) error {
	if _, e := s.repository.Assignment(c, id); e != nil {
		return e
	}
	if invalidAssignment(x) {
		return ErrInvalid
	}
	return dbErr(s.repository.UpdateAssignment(c, id, x))
}
func (s *Service) DeleteAssignment(c context.Context, id uint64) error {
	if _, e := s.repository.Assignment(c, id); e != nil {
		return e
	}
	return dbErr(s.repository.DeleteAssignment(c, id))
}
func invalidAssignment(x AssignmentInput) bool {
	return x.EmployeeID == 0 || x.PositionID == 0 || x.DepartmentID == 0 || x.StoreID == 0 || !validPeriod(x.StartedAt, x.EndedAt)
}

func (s *Service) CreateReportingLine(c context.Context, x ReportingLineInput) error {
	if e := s.validateReporting(c, x); e != nil {
		return e
	}
	return dbErr(s.repository.CreateReportingLine(c, x))
}
func (s *Service) UpdateReportingLine(c context.Context, id uint64, x ReportingLineInput) error {
	if _, e := s.repository.ReportingLine(c, id); e != nil {
		return e
	}
	if e := s.validateReporting(c, x); e != nil {
		return e
	}
	return dbErr(s.repository.UpdateReportingLine(c, id, x))
}
func (s *Service) DeleteReportingLine(c context.Context, id uint64) error {
	return dbErr(s.repository.DeleteReportingLine(c, id))
}
func (s *Service) validateReporting(c context.Context, x ReportingLineInput) error {
	if x.EmployeeAssignmentID == 0 || x.SupervisorAssignmentID == 0 || x.EmployeeAssignmentID == x.SupervisorAssignmentID || !validPeriod(x.StartedAt, x.EndedAt) {
		return ErrInvalid
	}
	a, e := s.repository.Assignment(c, x.EmployeeAssignmentID)
	if e != nil {
		return ErrInvalid
	}
	b, e := s.repository.Assignment(c, x.SupervisorAssignmentID)
	if e != nil || a.EmployeeID == b.EmployeeID {
		return ErrInvalid
	}
	return nil
}

func validPeriod(start, end string) bool {
	if !validDate(start, false) || !validDate(end, true) {
		return false
	}
	if end == "" {
		return true
	}
	return end >= start
}
func validDate(v string, optional bool) bool {
	if v == "" {
		return optional
	}
	_, e := time.Parse("2006-01-02", v)
	return e == nil
}
func oneOf(v string, values ...string) bool {
	for _, x := range values {
		if v == x {
			return true
		}
	}
	return false
}
func dbErr(e error) error {
	var me *mysql.MySQLError
	if errors.As(e, &me) {
		switch me.Number {
		case 1062:
			return ErrDuplicate
		case 1451, 1452:
			return ErrInUse
		case 3819, 4025:
			return ErrInvalid
		}
	}
	return e
}

package employee

import "database/sql"

type Employee struct {
	ID                                                        uint64
	EmployeeNumber, Name, Email, Phone, Status                string
	EmploymentStatus, PositionName, DepartmentName, StoreName string
	HasProfile, HasCurrentAssignment, HasUser                 bool
	AssignmentCount, SubordinateCount                         int
}

type Profile struct {
	EmployeeID                                                                           uint64
	EmployeeNumber, EmployeeName                                                         string
	BirthPlace, BirthDate, Gender, Address, JoinedDate, EmploymentStatus                 string
	PhotoPath, EmergencyContactName, EmergencyContactPhone, EmergencyContactRelationship string
}

type Assignment struct {
	ID, EmployeeID, PositionID, DepartmentID, StoreID                                   uint64
	EmployeeNumber, EmployeeName, PositionName, JobLevelName, DepartmentName, StoreName string
	StartedAt, EndedAt                                                                  string
	Current                                                                             bool
}

type ReportingLine struct {
	ID, EmployeeAssignmentID, SupervisorAssignmentID     uint64
	EmployeeNumber, EmployeeName, EmployeePosition       string
	SupervisorNumber, SupervisorName, SupervisorPosition string
	StartedAt, EndedAt                                   string
	Current                                              bool
}

type Option struct {
	ID    uint64
	Label string
}

type EmployeeInput struct{ EmployeeNumber, Name, Email, Phone, Status string }
type ProfileInput struct {
	EmployeeID                                                                           uint64
	BirthPlace, BirthDate, Gender, Address, JoinedDate, EmploymentStatus                 string
	PhotoPath, EmergencyContactName, EmergencyContactPhone, EmergencyContactRelationship string
}
type AssignmentInput struct {
	EmployeeID, PositionID, DepartmentID, StoreID uint64
	StartedAt, EndedAt                            string
}
type ReportingLineInput struct {
	EmployeeAssignmentID, SupervisorAssignmentID uint64
	StartedAt, EndedAt                           string
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func scanString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

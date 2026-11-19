package assessmentrubric

type Rubric struct {
	ID, PositionID, DepartmentID             uint64
	PositionName, DepartmentName, Code, Name string
	MonthFrom, MonthTo                       uint16
	IsActive                                 bool
	RuleCount, AssessmentCount               int
}
type Input struct {
	PositionID         uint64
	MonthFrom, MonthTo uint16
	IsActive           bool
}
type Position struct {
	ID   uint64
	Name string
}

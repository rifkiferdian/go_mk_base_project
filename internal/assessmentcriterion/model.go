package assessmentcriterion

type Criterion struct {
	DepartmentID      uint64
	DepartmentName    string
	MonthLabel        string
	ID                uint64
	AspectID          uint64
	AspectCode        string
	AspectName        string
	Code              string
	Description       string
	Weight            float64
	SortOrder         uint16
	RuleCount         int
	ExpectedRuleCount int
	AssessmentCount   int
}

type AspectOption struct {
	ID   uint64
	Name string
}

type RubricOption struct {
	ID             uint64
	DepartmentName string
	Name           string
	MonthLabel     string
}

type Input struct {
	DepartmentID uint64
	AspectID     uint64
	Code         string
	Description  string
	Weight       float64
	SortOrder    uint16
}

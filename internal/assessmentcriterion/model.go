package assessmentcriterion

func (c Criterion) EmploymentStatusLabel() string {
	switch c.EmploymentStatus {
	case "trainee":
		return "Trainee"
	case "probation":
		return "Probation"
	case "contract":
		return "Kontrak"
	case "permanent":
		return "Tetap"
	case "daily_worker":
		return "HL"
	case "post_retirement_daily_worker":
		return "HLPP"
	default:
		return c.EmploymentStatus
	}
}

type Criterion struct {
	EmploymentStatus  string
	DepartmentID      uint64
	DepartmentName    string
	PositionIDs       string
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
	EmploymentStatus string
	DepartmentID     uint64
	PositionIDs      []uint64
	AspectID         uint64
	Code             string
	Description      string
	Weight           float64
	SortOrder        uint16
}

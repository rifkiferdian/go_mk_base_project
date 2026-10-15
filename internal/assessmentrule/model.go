package assessmentrule

type Rule struct {
	RubricID             uint64
	RubricName           string
	DepartmentName       string
	MonthLabel           string
	CriterionID          uint64
	CriterionCode        string
	CriterionDescription string
	AspectID             uint64
	AspectName           string
	Score                uint8
	FindingMin           uint16
	FindingMax           uint16
	HasFindingMax        bool
	Description          string
}

type RuleGroup struct {
	RubricID, CriterionID, AspectID, DepartmentID uint64
	DepartmentName, MonthLabel, CriterionCode     string
	CriterionDescription, AspectName              string
	Rules                                         []Rule
}

type CriterionOption struct {
	ID                uint64
	AspectID          uint64
	AspectName        string
	Code              string
	Description       string
	RuleCount         int
	ExpectedRuleCount int
	DepartmentID      uint64
	DepartmentName    string
}

type AspectOption struct {
	ID   uint64
	Name string
}
type RubricOption struct {
	ID, DepartmentID                 uint64
	Name, DepartmentName, MonthLabel string
}

type Input struct {
	RubricID    uint64
	CriterionID uint64
	Score       uint8
	FindingMin  uint16
	FindingMax  *uint16
	Description string
}

type Key struct {
	RubricID    uint64
	CriterionID uint64
	Score       uint8
}

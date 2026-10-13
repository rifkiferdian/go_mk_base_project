package assessmentrule

type Rule struct {
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

type CriterionOption struct {
	ID          uint64
	AspectID    uint64
	AspectName  string
	Code        string
	Description string
	RuleCount   int
}

type AspectOption struct {
	ID   uint64
	Name string
}

type Input struct {
	CriterionID uint64
	Score       uint8
	FindingMin  uint16
	FindingMax  *uint16
	Description string
}

type Key struct {
	CriterionID uint64
	Score       uint8
}

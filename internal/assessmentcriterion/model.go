package assessmentcriterion

type Criterion struct {
	ID              uint64
	AspectID        uint64
	AspectCode      string
	AspectName      string
	Code            string
	Description     string
	Weight          float64
	SortOrder       uint16
	RuleCount       int
	AssessmentCount int
}

type AspectOption struct {
	ID   uint64
	Name string
}

type Input struct {
	AspectID    uint64
	Code        string
	Description string
	Weight      float64
	SortOrder   uint16
}

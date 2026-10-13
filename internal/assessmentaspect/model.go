package assessmentaspect

type Aspect struct {
	ID             uint64
	Code           string
	Name           string
	SortOrder      uint16
	CriterionCount int
}

type Input struct {
	Code      string
	Name      string
	SortOrder uint16
}

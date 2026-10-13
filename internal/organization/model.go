package organization

type JobLevel struct {
	ID            uint64
	Code          string
	Name          string
	RankOrder     uint16
	PositionCount int
}

type Position struct {
	ID              uint64
	JobLevelID      uint64
	JobLevelName    string
	JobLevelRank    uint16
	Code            string
	Name            string
	AssignmentCount int
}

type Department struct {
	ID              uint64
	Code            string
	Name            string
	AssignmentCount int
}

type JobLevelInput struct {
	Code, Name string
	RankOrder  uint16
}
type PositionInput struct {
	JobLevelID uint64
	Code, Name string
}
type DepartmentInput struct{ Code, Name string }

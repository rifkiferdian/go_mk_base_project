package assessmentreport

type Summary struct {
	AssessmentCount int
	TraineeCount    int
	StoreCount      int
	AverageScore    float64
	HighestScore    float64
}

type Distribution struct {
	Label      string
	Count      int
	Percentage float64
	Offset     float64
	Color      string
}

type Performance struct {
	Name       string
	Count      int
	Average    float64
	Percentage float64
}

type RecentResult struct {
	EmployeeName   string
	EmployeeNumber string
	PositionName   string
	DepartmentName string
	StoreName      string
	TrainingMonth  int
	PeriodName     string
	AssessmentDate string
	AssessorName   string
	AverageScore   float64
	Rating         string
}

type Filter struct {
	DepartmentID uint64
	RubricID     uint64
}

type DepartmentOption struct {
	ID   uint64
	Name string
}

type RubricOption struct {
	ID           uint64
	DepartmentID uint64
	Name         string
	PeriodName   string
}

type Report struct {
	Summary      Summary
	Distribution []Distribution
	Aspects      []Performance
	Stores       []Performance
	Departments  []Performance
	Periods      []Performance
	Recent       []RecentResult
}

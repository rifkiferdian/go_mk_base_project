package traineeassessment

type Assessment struct {
	ID, TraineeEmployeeID, TrainingMonth                                  uint64
	EmployeeNumber, EmployeeName, PositionName, DepartmentName, StoreName string
	AssessmentDate, Notes, Status, AssessorName, AssessorPosition         string
	TotalScore, TotalWeight, AverageScore                                 float64
	Rating                                                                string
	ScoredCount, CriterionCount                                           int
	Details                                                               []AssessmentDetail
}
type AssessmentDetail struct {
	AspectName, CriterionDescription string
	Weight                           float64
	Score                            int
}
type Trainee struct {
	EmployeeID, AssignmentID, StoreID                             uint64
	EmployeeNumber, Name, PositionName, DepartmentName, StoreName string
}
type Rule struct {
	Score        int
	Description  string
	FindingLabel string
}
type Criterion struct {
	ID, AspectID            uint64
	AspectName, Description string
	Weight                  float64
	Score                   int
	Rules                   []Rule
}
type Aspect struct {
	ID       uint64
	Name     string
	Criteria []Criterion
}
type FormData struct {
	Assessment Assessment
	Trainees   []Trainee
	Aspects    []Aspect
	Editing    bool
}
type DetailInput struct {
	CriterionID uint64
	Score       int
}
type Input struct {
	TraineeEmployeeID, TrainingMonth uint64
	AssessmentDate, Notes, Status    string
	Details                          []DetailInput
}

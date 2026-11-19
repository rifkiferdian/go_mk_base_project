package assessmentcategory

type Category struct {
	ID                          uint64
	Name                        string
	Minimum, Maximum, Deduction float64
	SortOrder                   uint16
	IsActive                    bool
}

type Input struct {
	Name                        string
	Minimum, Maximum, Deduction float64
	SortOrder                   uint16
	IsActive                    bool
}

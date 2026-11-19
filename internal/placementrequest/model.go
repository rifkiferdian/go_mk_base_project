package placementrequest

type Request struct {
	ID, EmployeeID                                                   uint64
	EmployeeNumber, EmployeeName, CurrentPosition, RequestedPosition string
	CurrentStore, RequestedStore, EffectiveDate, Reason, Status      string
	RequesterName, DecisionNotes, DeciderName, CreatedAt             string
}
type Option struct {
	ID           uint64
	Name         string
	DepartmentID uint64
}
type Input struct {
	EmployeeID, PositionID, StoreID uint64
	EffectiveDate, Reason           string
}

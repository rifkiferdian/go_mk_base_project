package placementrequest

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalid = errors.New("invalid placement request")
	ErrFuture  = errors.New("effective date is in the future")
)

type Service struct{ r *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func all(roles string) bool {
	return strings.Contains(","+strings.ReplaceAll(roles, " ", "")+",", ",superadmin,") || strings.Contains(","+strings.ReplaceAll(roles, " ", "")+",", ",manager,")
}
func (s *Service) Page(ctx context.Context, userID uint64, roles string) ([]Request, []Option, []Option, []Option, error) {
	a := all(roles)
	items, e := s.r.Requests(ctx, userID, a)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	employees, e := s.r.Employees(ctx, userID, a)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	positions, e := s.r.Positions(ctx)
	if e != nil {
		return nil, nil, nil, nil, e
	}
	stores, e := s.r.Stores(ctx)
	return items, employees, positions, stores, e
}
func (s *Service) Create(ctx context.Context, userID uint64, roles string, x Input) error {
	x.Reason = strings.TrimSpace(x.Reason)
	if x.EmployeeID == 0 || x.PositionID == 0 || x.StoreID == 0 || x.Reason == "" {
		return ErrInvalid
	}
	if _, e := time.Parse("2006-01-02", x.EffectiveDate); e != nil {
		return ErrInvalid
	}
	return s.r.Create(ctx, userID, x, all(roles))
}
func (s *Service) Decide(ctx context.Context, id, userID uint64, approve bool, notes string) error {
	if id == 0 {
		return ErrInvalid
	}
	return s.r.Decide(ctx, id, userID, approve, strings.TrimSpace(notes))
}

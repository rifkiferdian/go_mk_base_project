package assessmentrubric

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid = errors.New("invalid rubric period")
	ErrOverlap = errors.New("rubric period overlaps")
	ErrInUse   = errors.New("rubric period in use")
)

type Service struct{ r *Repository }

func NewService(r *Repository) *Service { return &Service{r} }
func (s *Service) Page(ctx context.Context) ([]Rubric, []Position, error) {
	items, e := s.r.Rubrics(ctx)
	if e != nil {
		return nil, nil, e
	}
	p, e := s.r.Positions(ctx)
	return items, p, e
}
func (s *Service) Create(ctx context.Context, x Input) error {
	if e := s.validate(ctx, 0, x); e != nil {
		return e
	}
	return dbErr(s.r.Create(ctx, x))
}
func (s *Service) Update(ctx context.Context, id uint64, x Input) error {
	old, e := s.r.Rubric(ctx, id)
	if e != nil {
		return e
	}
	if old.AssessmentCount > 0 {
		return ErrInUse
	}
	if old.RuleCount > 0 && old.PositionID != x.PositionID {
		return ErrInUse
	}
	if e = s.validate(ctx, id, x); e != nil {
		return e
	}
	return dbErr(s.r.Update(ctx, id, x))
}
func (s *Service) Delete(ctx context.Context, id uint64) error {
	x, e := s.r.Rubric(ctx, id)
	if e != nil {
		return e
	}
	if x.RuleCount > 0 || x.AssessmentCount > 0 {
		return ErrInUse
	}
	return dbErr(s.r.Delete(ctx, id))
}
func (s *Service) validate(ctx context.Context, id uint64, x Input) error {
	if x.PositionID == 0 || x.MonthFrom == 0 || x.MonthTo < x.MonthFrom || x.MonthTo > 24 {
		return ErrInvalid
	}
	overlap, e := s.r.Overlaps(ctx, id, x.PositionID, x.MonthFrom, x.MonthTo)
	if e != nil {
		return e
	}
	if overlap {
		return ErrOverlap
	}
	return nil
}
func dbErr(e error) error {
	var me *mysql.MySQLError
	if errors.As(e, &me) && me.Number == 1062 {
		return ErrOverlap
	}
	if errors.As(e, &me) && me.Number == 1451 {
		return ErrInUse
	}
	return e
}

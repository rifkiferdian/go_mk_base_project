package assessmentcategory

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalid = errors.New("invalid assessment result category")
	ErrOverlap = errors.New("assessment result category range overlaps")
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) Categories(ctx context.Context) ([]Category, error) {
	return s.repository.Categories(ctx)
}
func (s *Service) Create(ctx context.Context, input Input) error { return s.save(ctx, 0, input) }
func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.Category(ctx, id); err != nil {
		return err
	}
	return s.save(ctx, id, input)
}
func (s *Service) Delete(ctx context.Context, id uint64) error { return s.repository.Delete(ctx, id) }
func (s *Service) save(ctx context.Context, id uint64, input Input) error {
	input.Name = strings.ToUpper(strings.TrimSpace(input.Name))
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 100 || input.Minimum < 1 || input.Maximum > 5 || input.Minimum > input.Maximum || input.Deduction < 0 || input.Deduction > 100 || input.SortOrder == 0 {
		return ErrInvalid
	}
	overlap, err := s.repository.Overlaps(ctx, id, input.Minimum, input.Maximum)
	if err != nil {
		return err
	}
	if overlap {
		return ErrOverlap
	}
	if id == 0 {
		return s.repository.Create(ctx, input)
	}
	return s.repository.Update(ctx, id, input)
}

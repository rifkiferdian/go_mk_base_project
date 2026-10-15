package assessmentcriterion

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid assessment criterion")
	ErrDuplicate = errors.New("duplicate assessment criterion")
	ErrInUse     = errors.New("assessment criterion in use")
)
var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,49}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) Criteria(ctx context.Context) ([]Criterion, error) {
	return s.repository.Criteria(ctx)
}
func (s *Service) Aspects(ctx context.Context) ([]AspectOption, error) {
	return s.repository.Aspects(ctx)
}
func (s *Service) Rubrics(ctx context.Context) ([]RubricOption, error) {
	return s.repository.Rubrics(ctx)
}
func (s *Service) Departments(ctx context.Context) ([]AspectOption, error) {
	return s.repository.Departments(ctx)
}
func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := s.normalize(ctx, input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Create(ctx, input))
}
func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	item, err := s.repository.Criterion(ctx, id)
	if err != nil {
		return err
	}
	if item.AssessmentCount > 0 {
		return ErrInUse
	}
	input, err = s.normalize(ctx, input)
	if err != nil {
		return err
	}
	return databaseError(s.repository.Update(ctx, id, input))
}
func (s *Service) Delete(ctx context.Context, id uint64) error {
	item, err := s.repository.Criterion(ctx, id)
	if err != nil {
		return err
	}
	if item.AssessmentCount > 0 {
		return ErrInUse
	}
	return databaseError(s.repository.Delete(ctx, id))
}

func (s *Service) normalize(ctx context.Context, input Input) (Input, error) {
	if input.DepartmentID == 0 {
		return Input{}, ErrInvalid
	}
	active, err := s.repository.DepartmentExists(ctx, input.DepartmentID)
	if err != nil {
		return Input{}, err
	}
	if !active {
		return Input{}, ErrInvalid
	}
	input.Code = strings.ToLower(strings.TrimSpace(input.Code))
	input.Description = strings.TrimSpace(input.Description)
	if input.AspectID == 0 || !codePattern.MatchString(input.Code) || input.Description == "" || utf8.RuneCountInString(input.Description) > 2000 || input.Weight <= 0 || input.Weight > 999999.99 || math.IsNaN(input.Weight) || math.IsInf(input.Weight, 0) || input.SortOrder == 0 {
		return Input{}, ErrInvalid
	}
	exists, err := s.repository.AspectExists(ctx, input.AspectID)
	if err != nil {
		return Input{}, err
	}
	if !exists {
		return Input{}, ErrInvalid
	}
	return input, nil
}
func databaseError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrDuplicate
		case 1451, 1452:
			return ErrInUse
		}
	}
	return err
}

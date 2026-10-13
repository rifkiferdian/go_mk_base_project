package assessmentaspect

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid assessment aspect")
	ErrDuplicate = errors.New("duplicate assessment aspect")
	ErrInUse     = errors.New("assessment aspect in use")
)

var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,49}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service                 { return &Service{repository: repository} }
func (s *Service) Aspects(ctx context.Context) ([]Aspect, error) { return s.repository.Aspects(ctx) }

func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return translateDatabaseError(s.repository.Create(ctx, input))
}

func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.Aspect(ctx, id); err != nil {
		return err
	}
	input, err := normalize(input)
	if err != nil {
		return err
	}
	return translateDatabaseError(s.repository.Update(ctx, id, input))
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	item, err := s.repository.Aspect(ctx, id)
	if err != nil {
		return err
	}
	if item.CriterionCount > 0 {
		return ErrInUse
	}
	return translateDatabaseError(s.repository.Delete(ctx, id))
}

func normalize(input Input) (Input, error) {
	input.Code = strings.ToLower(strings.TrimSpace(input.Code))
	input.Name = strings.ToUpper(strings.TrimSpace(input.Name))
	if !codePattern.MatchString(input.Code) || input.Name == "" || utf8.RuneCountInString(input.Name) > 100 || input.SortOrder == 0 {
		return Input{}, ErrInvalid
	}
	return input, nil
}

func translateDatabaseError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return ErrDuplicate
		case 1451:
			return ErrInUse
		}
	}
	return err
}

package organization

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid organization master")
	ErrDuplicate = errors.New("duplicate organization master")
	ErrInUse     = errors.New("organization master in use")
)
var codePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,49}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) JobLevels(ctx context.Context) ([]JobLevel, error) {
	return s.repository.JobLevels(ctx)
}
func (s *Service) Positions(ctx context.Context) ([]Position, error) {
	return s.repository.Positions(ctx)
}
func (s *Service) Departments(ctx context.Context) ([]Department, error) {
	return s.repository.Departments(ctx)
}

func (s *Service) CreateJobLevel(ctx context.Context, input JobLevelInput) error {
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if invalidCodeName(input.Code, input.Name) || input.RankOrder == 0 {
		return ErrInvalid
	}
	return dbError(s.repository.CreateJobLevel(ctx, input))
}
func (s *Service) UpdateJobLevel(ctx context.Context, id uint64, input JobLevelInput) error {
	if _, err := s.repository.JobLevel(ctx, id); err != nil {
		return err
	}
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if invalidCodeName(input.Code, input.Name) || input.RankOrder == 0 {
		return ErrInvalid
	}
	return dbError(s.repository.UpdateJobLevel(ctx, id, input))
}
func (s *Service) DeleteJobLevel(ctx context.Context, id uint64) error {
	item, err := s.repository.JobLevel(ctx, id)
	if err != nil {
		return err
	}
	if item.PositionCount > 0 {
		return ErrInUse
	}
	return dbError(s.repository.DeleteJobLevel(ctx, id))
}

func (s *Service) CreatePosition(ctx context.Context, input PositionInput) error {
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if input.JobLevelID == 0 || invalidCodeName(input.Code, input.Name) {
		return ErrInvalid
	}
	if _, err := s.repository.JobLevel(ctx, input.JobLevelID); err != nil {
		return ErrInvalid
	}
	return dbError(s.repository.CreatePosition(ctx, input))
}
func (s *Service) UpdatePosition(ctx context.Context, id uint64, input PositionInput) error {
	if _, err := s.repository.Position(ctx, id); err != nil {
		return err
	}
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if input.JobLevelID == 0 || invalidCodeName(input.Code, input.Name) {
		return ErrInvalid
	}
	if _, err := s.repository.JobLevel(ctx, input.JobLevelID); err != nil {
		return ErrInvalid
	}
	return dbError(s.repository.UpdatePosition(ctx, id, input))
}
func (s *Service) DeletePosition(ctx context.Context, id uint64) error {
	item, err := s.repository.Position(ctx, id)
	if err != nil {
		return err
	}
	if item.AssignmentCount > 0 {
		return ErrInUse
	}
	return dbError(s.repository.DeletePosition(ctx, id))
}

func (s *Service) CreateDepartment(ctx context.Context, input DepartmentInput) error {
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if invalidCodeName(input.Code, input.Name) {
		return ErrInvalid
	}
	return dbError(s.repository.CreateDepartment(ctx, input))
}
func (s *Service) UpdateDepartment(ctx context.Context, id uint64, input DepartmentInput) error {
	if _, err := s.repository.Department(ctx, id); err != nil {
		return err
	}
	input.Code, input.Name = normalizeText(input.Code, input.Name)
	if invalidCodeName(input.Code, input.Name) {
		return ErrInvalid
	}
	return dbError(s.repository.UpdateDepartment(ctx, id, input))
}
func (s *Service) DeleteDepartment(ctx context.Context, id uint64) error {
	item, err := s.repository.Department(ctx, id)
	if err != nil {
		return err
	}
	if item.AssignmentCount > 0 {
		return ErrInUse
	}
	return dbError(s.repository.DeleteDepartment(ctx, id))
}

func normalizeText(code, name string) (string, string) {
	return strings.ToLower(strings.TrimSpace(code)), strings.TrimSpace(name)
}
func invalidCodeName(code, name string) bool {
	return !codePattern.MatchString(code) || name == "" || utf8.RuneCountInString(name) > 150
}
func dbError(err error) error {
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

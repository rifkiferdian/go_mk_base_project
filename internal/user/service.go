package user

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalid   = errors.New("invalid user")
	ErrDuplicate = errors.New("duplicate user")
	ErrProtected = errors.New("protected user")
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,99}$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service             { return &Service{repository: repository} }
func (s *Service) Users(ctx context.Context) ([]User, error) { return s.repository.Users(ctx) }
func (s *Service) Options(ctx context.Context) ([]StoreOption, []RoleOption, error) {
	return s.repository.Options(ctx)
}
func (s *Service) Employees(ctx context.Context) ([]EmployeeOption, error) {
	return s.repository.Employees(ctx)
}
func (s *Service) StoreAssignments(ctx context.Context, userID uint64) (User, []StoreAssignment, []RoleOption, error) {
	item, err := s.repository.User(ctx, userID)
	if err != nil {
		return User{}, nil, nil, err
	}
	assignments, err := s.repository.StoreAssignments(ctx, userID)
	if err != nil {
		return User{}, nil, nil, err
	}
	_, roles, err := s.repository.Options(ctx)
	return item, assignments, roles, err
}
func (s *Service) SyncStoreAssignments(ctx context.Context, userID uint64, assignments []AssignmentInput) error {
	if _, err := s.repository.User(ctx, userID); err != nil {
		return err
	}
	if len(assignments) == 0 {
		return ErrInvalid
	}
	seen := make(map[uint64]bool, len(assignments))
	defaultCount := 0
	for _, item := range assignments {
		if item.StoreID == 0 || item.RoleID == 0 || seen[item.StoreID] {
			return ErrInvalid
		}
		seen[item.StoreID] = true
		if item.IsDefault {
			defaultCount++
		}
	}
	if defaultCount != 1 {
		return ErrInvalid
	}
	return s.repository.SyncStoreAssignments(ctx, userID, assignments)
}

func (s *Service) Create(ctx context.Context, input Input) error {
	input, err := normalize(input, true)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return duplicate(s.repository.Create(ctx, input, string(hash)))
}
func (s *Service) Update(ctx context.Context, id uint64, input Input) error {
	if _, err := s.repository.User(ctx, id); err != nil {
		return err
	}
	input, err := normalize(input, false)
	if err != nil {
		return err
	}
	hash := ""
	if input.Password != "" {
		value, e := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if e != nil {
			return e
		}
		hash = string(value)
	}
	return duplicate(s.repository.Update(ctx, id, input, hash))
}
func (s *Service) Delete(ctx context.Context, id, currentID uint64) error {
	if id == currentID {
		return ErrProtected
	}
	if _, err := s.repository.User(ctx, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func normalize(input Input, passwordRequired bool) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Username = strings.ToLower(strings.TrimSpace(input.Username))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.EmployeeNumber = strings.TrimSpace(input.EmployeeNumber)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Status != "active" && input.Status != "inactive" {
		return Input{}, ErrInvalid
	}
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 150 || !usernamePattern.MatchString(input.Username) || len(input.EmployeeNumber) > 50 || len(input.Phone) > 30 || input.StoreID == 0 || input.RoleID == 0 {
		return Input{}, ErrInvalid
	}
	if input.Email != "" {
		address, err := mail.ParseAddress(input.Email)
		if err != nil || address.Address != input.Email || len(input.Email) > 191 {
			return Input{}, ErrInvalid
		}
	}
	if passwordRequired && input.Password == "" {
		return Input{}, ErrInvalid
	}
	if input.Password != "" && (len(input.Password) < 8 || len(input.Password) > 72) {
		return Input{}, ErrInvalid
	}
	return input, nil
}
func duplicate(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicate
	}
	return err
}

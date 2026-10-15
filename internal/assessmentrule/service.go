package assessmentrule

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalid   = errors.New("invalid assessment score rule")
	ErrDuplicate = errors.New("duplicate assessment score rule")
	ErrOverlap   = errors.New("overlapping assessment score rule")
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service             { return &Service{repository: repository} }
func (s *Service) Rules(ctx context.Context) ([]Rule, error) { return s.repository.Rules(ctx) }
func (s *Service) Criteria(ctx context.Context) ([]CriterionOption, error) {
	return s.repository.Criteria(ctx)
}
func (s *Service) Rubrics(ctx context.Context) ([]RubricOption, error) {
	return s.repository.Rubrics(ctx)
}
func (s *Service) Create(ctx context.Context, input Input) error {
	if err := s.validate(ctx, input, nil); err != nil {
		return err
	}
	return databaseError(s.repository.Create(ctx, input))
}
func (s *Service) Update(ctx context.Context, old Key, input Input) error {
	if _, err := s.repository.Rule(ctx, old); err != nil {
		return err
	}
	if err := s.validate(ctx, input, &old); err != nil {
		return err
	}
	return databaseError(s.repository.Update(ctx, old, input))
}
func (s *Service) Delete(ctx context.Context, key Key) error {
	if _, err := s.repository.Rule(ctx, key); err != nil {
		return err
	}
	return s.repository.Delete(ctx, key)
}
func (s *Service) validate(ctx context.Context, input Input, exclude *Key) error {
	input.Description = strings.TrimSpace(input.Description)
	if input.RubricID == 0 || input.CriterionID == 0 || input.Score < 1 || input.Score > 5 || input.Description == "" || utf8.RuneCountInString(input.Description) > 150 || input.FindingMax != nil && *input.FindingMax < input.FindingMin {
		return ErrInvalid
	}
	exists, err := s.repository.CriterionMatchesRubric(ctx, input.CriterionID, input.RubricID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalid
	}
	rules, err := s.repository.RulesForCriterion(ctx, input.RubricID, input.CriterionID)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if exclude != nil && rule.RubricID == exclude.RubricID && rule.CriterionID == exclude.CriterionID && rule.Score == exclude.Score {
			continue
		}
		if rangesOverlap(input, rule) {
			return ErrOverlap
		}
	}
	return nil
}
func rangesOverlap(input Input, rule Rule) bool {
	inputMax := uint32(65536)
	if input.FindingMax != nil {
		inputMax = uint32(*input.FindingMax)
	}
	ruleMax := uint32(65536)
	if rule.HasFindingMax {
		ruleMax = uint32(rule.FindingMax)
	}
	return uint32(input.FindingMin) <= ruleMax && uint32(rule.FindingMin) <= inputMax
}
func databaseError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicate
	}
	return err
}
